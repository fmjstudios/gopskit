package util

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fmjstudios/gopskit/internal/ssolo/app"
	"github.com/fmjstudios/gopskit/pkg/core"
	fs "github.com/fmjstudios/gopskit/pkg/fsi"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Realm    string `json:"realm"`
}

type Credentials struct {
	Hosts  map[string]Login
	Tokens map[string]string
}

// CredentialPath builds the filesystem path to write the credentials to after we unseal the Vault,
// since it most likely is required for later commands. This function make the path deterministic
// per execution env.Environment.
func CredentialPath(a *app.State, env core.Environment) string {
	return filepath.Join(a.Paths.Cache, env.String(), "keycloak-credentials.json")
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

// Pods returns a list of Kubernetes' Pods matching the default (or custom) Keycloak label
// TODO(FMJdev): this is duplicated because the original source was a internal package
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

// TODO(FMJdev): this is duplicated because the original source was a internal package
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

// TODO(FMJdev): this is duplicated because the original source was a internal package
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
