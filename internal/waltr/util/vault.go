package util

import (
	"context"
	json "encoding/json"
	"fmt"
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/pkg/env"
	"github.com/fmjstudios/gopskit/pkg/filesystem"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	"strings"
)

// Credentials is a custom type which is used to write and load Vault credentials to and from a file
type Credentials struct {
	Keys       []string `json:"keys"`
	KeysBase64 []string `json:"keys_base64"`
	Token      string   `json:"token"`
}

// CredentialPath builds the filesystem path to write the credentials to after we unseal the Vault,
// since it most likely is required for later commands. This function make the path deterministic
// per execution env.Environment.
func CredentialPath(a *app.App, env env.Environment) string {
	return filepath.Join(a.Platform.CacheDir(), env.String(), "vault-credentials.json")
}

// WriteCredentials writes the Vault Credentials to the CredentialPath for the given env.Environment
func WriteCredentials(a *app.App, env env.Environment, credentials *Credentials) error {
	p := CredentialPath(a, env)
	jsn, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return err
	}

	return filesystem.Write(p, jsn)
}

// ReadCredentials reads the Vault Credentials from the CredentialPath for the given env.Environment
func ReadCredentials(a *app.App, env env.Environment) (*Credentials, error) {
	p := CredentialPath(a, env)
	raw, err := filesystem.Read(p)
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
func AuthMethods(a *app.App) ([]string, error) {
	// get current methods
	m, err := a.VaultClient.System.AuthListEnabledMethods(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not list enabled Vault authentication methods: %v", err)
	}

	var methods []string
	for k, _ := range m.Data {
		methods = append(methods, k)
		continue
	}

	return methods, nil
}

// SecretsEngines retrieves the list of enabled secrets engines from the current Vault instance
func SecretsEngines(a *app.App) ([]string, error) {
	// get current secrets engines
	s, err := a.VaultClient.System.MountsListSecretsEngines(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not list secrets engines: %v", err)
	}

	var engines []string
	for k, _ := range s.Data {
		engines = append(engines, k)
		continue
	}

	return engines, nil
}

// Policies retrieves a list of the currently enabled policies
func Policies(a *app.App) ([]string, error) {
	p, err := a.VaultClient.System.PoliciesListAclPolicies(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not list policies: %v", err)
	}

	return p.Data.Keys, nil
}

// PasswordPolicies retrieves a list of the currently enabled password policies
func PasswordPolicies(a *app.App) ([]string, error) {
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
func KubernetesAuthRoles(a *app.App) ([]string, error) {
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
func GeneratePasswordFromPolicy(a *app.App, policy string) (string, error) {
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
func Pods(a *app.App, namespace, label string) ([]corev1.Pod, error) {
	if label == "" {
		label = app.DefaultLabel
	}

	pods, err := a.KubeClient.Pods(namespace, metav1.ListOptions{
		LabelSelector: label,
	})

	if err != nil {
		return nil, err
	}

	return pods, nil
}

// Contains checks if an element is contained within a string slice
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
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
