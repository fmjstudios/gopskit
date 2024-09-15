package util

import (
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/pkg/env"
	"github.com/fmjstudios/gopskit/pkg/filesystem"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
)

// TokenPath builds the filesystem path to write the token to after we unseal the Vault,
// since it most likely is required for later commands. This function make the path deterministic
// per execution env.Environment.
func TokenPath(a *app.App, env env.Environment) string {
	return filepath.Join(a.Platform.CacheDir(), env.String(), "vault", "vault-token")
}

// WriteToken writes the Vault Token to the TokenPath for the given env.Environment
func WriteToken(a *app.App, env env.Environment, token string) error {
	p := TokenPath(a, env)
	if err := filesystem.Write(p, []byte(token)); err != nil {
		return err
	}

	return nil
}

// ReadToken reads the Vault Token from the TokenPath for the given env.Environment
func ReadToken(a *app.App, env env.Environment) (string, error) {
	p := TokenPath(a, env)
	token, err := filesystem.Read(p)
	if err != nil {
		return "", err
	}

	return string(token), nil
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
