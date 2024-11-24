package util

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/pkg/core"
	fs "github.com/fmjstudios/gopskit/pkg/fsi"
	"github.com/fmjstudios/gopskit/pkg/helpers"
	"github.com/hashicorp/hcl/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VaultConfig struct {
	DisableMLock bool        `hcl:"disable_mlock"`
	UI           bool        `hcl:"ui"`
	Seal         *SealConfig `hcl:"seal,block"`
	Remain       hcl.Body    `hcl:",remain"`
}

type SealConfig struct {
	Type   string   `hcl:"type,label"`
	Remain hcl.Body `hcl:",remain"`
}

// Credentials is a custom type which is used to write and load Vault credentials to and from a file
type Credentials struct {
	Keys       []string `json:"keys"`
	KeysBase64 []string `json:"keys_base64"`
	Token      string   `json:"token"`
}

// CredentialPath builds the filesystem path to write the credentials to after we unseal the Vault,
// since it most likely is required for later commands. This function make the path deterministic
// per execution env.Environment.
func CredentialPath(a *app.State, env core.Environment) string {
	return filepath.Join(a.Paths.Cache, env.String(), "vault-credentials.json")
}

// WriteCredentials writes the Vault Credentials to the CredentialPath for the given env.Environment
func WriteCredentials(a *app.State, env core.Environment, credentials *Credentials) error {
	p := CredentialPath(a, env)
	jsn, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return err
	}

	return fs.Write(p, jsn)
}

// ReadCredentials reads the Vault Credentials from the CredentialPath for the given env.Environment
func ReadCredentials(a *app.State, env core.Environment) (*Credentials, error) {
	p := CredentialPath(a, env)
	raw, err := fs.Read(p)
	if err != nil {
		return nil, err
	}

	var credentials Credentials
	err = json.Unmarshal(raw, &credentials)
	if err != nil {
		return nil, err
	}

	return &credentials, nil
}

// AuthMethods retrieves the list of enabled authentication methods from the current Vault instance
func AuthMethods(a *app.State) ([]string, error) {
	// get current methods
	m, err := a.VaultClient.System.AuthListEnabledMethods(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not list enabled Vault authentication methods: %v", err)
	}

	var methods []string
	for k := range m.Data {
		methods = append(methods, k)
		continue
	}

	return methods, nil
}

// SecretsEngines retrieves the list of enabled secrets engines from the current Vault instance
func SecretsEngines(a *app.State) ([]string, error) {
	// get current secrets engines
	s, err := a.VaultClient.System.MountsListSecretsEngines(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not list secrets engines: %v", err)
	}

	var engines []string
	for k := range s.Data {
		engines = append(engines, k)
		continue
	}

	return engines, nil
}

// Policies retrieves a list of the currently enabled policies
func Policies(a *app.State) ([]string, error) {
	p, err := a.VaultClient.System.PoliciesListAclPolicies(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not list policies: %v", err)
	}

	return p.Data.Keys, nil
}

// PasswordPolicies retrieves a list of the currently enabled password policies
func PasswordPolicies(a *app.State) ([]string, error) {
	p, err := a.VaultClient.System.PoliciesListPasswordPolicies(context.Background())
	if err != nil {
		// mitigate empty policies
		if strings.Contains(err.Error(), "404 Not Found") {
			return []string{}, nil
		}

		return nil, fmt.Errorf("could not list password policies: %v", err)
	}

	return p.Data.Keys, nil
}

// KubernetesAuthRoles retrieves a list of the currently enabled Kubernetes Auth roles within Vault
func KubernetesAuthRoles(a *app.State) ([]string, error) {
	k, err := a.VaultClient.Auth.KubernetesListAuthRoles(context.Background())
	if err != nil {
		// mitigate empty policies
		if strings.Contains(err.Error(), "404 Not Found") {
			return []string{}, nil
		}

		return nil, fmt.Errorf("could not list Kubernetes Auth roles: %v", err)
	}

	return k.Data.Keys, nil
}

// GeneratePasswordFromPolicy ...
func GeneratePasswordFromPolicy(a *app.State, policy string) (string, error) {
	pass, err := a.VaultClient.System.PoliciesGeneratePasswordFromPasswordPolicy(
		context.Background(),
		policy,
	)
	if err != nil {
		return "", fmt.Errorf("could not generate password: %v", err)
	}
	return pass.Data.Password, nil
}

// Pods returns a list of Kubernetes' Pods matching the default (or custom) Vault label
func Pods(a *app.State, namespace, label string) ([]corev1.Pod, error) {
	if label == "" {
		a.Log.Debugf("no label provided for Pods search, using default label: %s", app.DefaultLabel)
		label = app.DefaultLabel
	}

	pods, err := a.Kube.Pods(namespace, metav1.ListOptions{
		LabelSelector: label,
	})
	a.Log.Debugf("found %d pods for label: %s", len(pods), label)

	if err != nil {
		return nil, err
	}

	return pods, nil
}

func LeaderPod(a *app.State, pods []corev1.Pod, namespace, label string) (*corev1.Pod, error) {
	var leader *corev1.Pod
	var activePods []corev1.Pod
	var err error

	if len(pods) > 1 {
		label = fmt.Sprintf("%s,%s", label, "vault-active=true")
		activePods, err = a.Kube.Pods(namespace, metav1.ListOptions{
			LabelSelector: label,
		})

		if err != nil {
			return nil, err
		}
	}

	// matched more than one
	if len(activePods) > 1 {
		return nil, fmt.Errorf("could not determine Vault leader pod. "+
			"Invalid configuration: Kubernetes label %s matched more than one pod", label)
	}

	// no pod is labeled with "vault-active=true"
	if len(activePods) == 0 {
		for _, pod := range pods {
			if strings.Contains(pod.Name, "0") {
				leader = &pod
			}
		}

		if leader == nil {
			return nil, fmt.Errorf("could not determine Vault leader pod. Unfamiliar naming scheme. " +
				"None of your Vault Pod names contain a zero")
		}

		return leader, nil
	}

	return &pods[0], nil
}

// EnsureNamespace ensures we only find and use Vault Pods within a single namespace
func EnsureNamespace(pods []corev1.Pod) (string, error) {
	var ns []string
	for _, pod := range pods {
		ns = append(ns, pod.Namespace)
	}

	rns := helpers.RemoveDuplicates(ns)
	if len(rns) > 1 {
		return "", fmt.Errorf("discovered Vault pods in multiple namespaces: %v! Please set the namespace option", rns)
	}

	return ns[0], nil
}

func WaitUntilRunning(a *app.State, pod corev1.Pod) {
	for {
		if pod.Status.Phase != corev1.PodRunning {
			a.Log.Infof("main vault Pod: %s is not running yet - waiting for Pod to start", pod.Name)
			time.Sleep(2500 * time.Millisecond)
		} else {
			a.Log.Infof("main vault Pod: %s is running", pod.Name)
			break
		}
	}
}

// Policies
var (
	Releases = []string{
		"keycloak",
		"awx",
		"crowdsec",
		"gitlab",
		"gitlab-runner",
		"harbor",
		"headlamp",
		"homepage",
		"jenkins",
		"kubescape",
		"loki",
		"matomo",
	}

	ConfigReleasePolicyTemplate = `path "kv/data/%s/*" {
   capabilities = ["read"]
}`

	ConfigAclPolicies = map[string]string{
		"admin": `# Read system health check
path "sys/health"
{
  capabilities = ["read", "sudo"]
}

# Create and manage ACL policies broadly across Vault

# List existing policies
path "sys/policies/acl"
{
  capabilities = ["list"]
}

# Create and manage ACL policies
path "sys/policies/acl/*"
{
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}

# Enable and manage authentication methods broadly across Vault

# Manage auth methods broadly across Vault
path "auth/*"
{
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}

# Create, update, and delete auth methods
path "sys/auth/*"
{
  capabilities = ["create", "update", "delete", "sudo"]
}

# List auth methods
path "sys/auth"
{
  capabilities = ["read"]
}

# Enable and manage the key/value secrets engine at \"secret\" path

# List, create, update, and delete key/value secrets
path "kv/*"
{
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}

# List, create, update, and delete transit secrets
path "transit/*"
{
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}

# List, create, update, and delete cubbyhole secrets
path "cubbyhole/*"
{
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}

# List, create, update, and delete Vault identities
path "identity/*"
{
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}

# Manage secrets engines
path "sys/mounts/*"
{
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}

# List existing secrets engines.
path "sys/mounts"
{
  capabilities = ["read"]
}`,
	}

	ConfigPasswordPolicies = map[string]string{
		"alphanumeric-password": `length = 64
rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
  min-chars = 16
}
rule "charset" {
  charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  min-chars = 16
}
rule "charset" {
  charset = "0123456789"
  min-chars = 16
}`,
		"alphanumeric-special-password": `length = 64
rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
  min-chars = 12
}
rule "charset" {
  charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  min-chars = 12
}
rule "charset" {
  charset = "0123456789"
  min-chars = 12
}
rule "charset" {
  charset = "!@#$%^&*"
  min-chars = 12
}`,
		"s3-access-key": `length = 32
rule "charset" {
  charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  min-chars = 16
}
rule "charset" {
  charset = "0123456789"
  min-chars = 12
}`,
		"s3-secret-key": `length = 64
rule "charset" {
  charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  min-chars = 32
}
rule "charset" {
  charset = "0123456789"
  min-chars = 24
}`,
	}

	TRANSIT_ENCRYPTION_POLICY = `path "transit/encrypt/vso-client-cache" {
   capabilities = ["create", "update"]
}
path "transit/decrypt/vso-client-cache" {
   capabilities = ["create", "update"]
}`
)
