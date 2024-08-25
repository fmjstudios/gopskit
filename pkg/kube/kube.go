package kube

import (
	"errors"
	"path/filepath"

	"github.com/fmjstudios/gopskit/pkg/filesystem"
	"github.com/fmjstudios/gopskit/pkg/platform"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	DefaultNamespace = "default"
)

var (
	searchPaths []string = []string{
		filepath.Join(platform.Current().Home(), ".kube", "config"),
		filepath.Join(platform.Current().ConfigDir(), "gopskit", "kubeconfig"),
	}
)

// ref: https://github.com/iximiuz/client-go-examples/blob/main/kubeconfig-from-yaml/main.go
// ref: https://github.com/a4abhishek/Client-Go-Examples/blob/master/exec_to_pod/exec_to_pod.go
// ref: https://miminar.fedorapeople.org/_preview/openshift-enterprise/registry-redeploy/go_client/executing_remote_processes.html
// ref: https://github.com/gianarb/kube-port-forward/blob/master/main.go
// ref: https://github.com/anthhub/forwarder

type KubeClient struct {
	// config is the configuration file path for which the current client(-set) was created
	config string

	// namespace is the Kubernetes namespace the client is configured to access
	namespace string

	// krc is the embedded rest.Config for which the client was built
	krc *rest.Config

	// cs is the embedded Kubernetes ClientSet
	kcs *kubernetes.Clientset
}

// Opt represents a configuration option for the KubeClient
type Opt func(c *KubeClient)

// NewClient instantiates a new KubeClient with a configured and embedded Clientset
func NewClient(opts ...Opt) (*KubeClient, error) {
	var err error

	kc := &KubeClient{}
	for _, opt := range opts {
		opt(kc)
	}

	if kc.config == "" {
		kc.config, err = findKubeConfig()
		if err != nil {
			return nil, err
		}
	}

	if kc.namespace == "" {
		kc.namespace = DefaultNamespace
	}

	conf, err := clientcmd.BuildConfigFromFlags("", kc.config)
	if err != nil {
		return nil, err
	}

	kc.kcs, err = kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	return kc, nil
}

// WithConfigPath configures the KubeClient with a predetermined path for the 'kubeconfig' file
// This avoids the usual searches done within findKubeConfig
func WithConfigPath(path string) func(c *KubeClient) {
	return func(c *KubeClient) {
		c.config = path
	}
}

// WithNamespace configures the KubeClient with a custom default namespace
func WithNamespace(namespace string) func(c *KubeClient) {
	return func(c *KubeClient) {
		c.namespace = namespace
	}
}

// TODO(FMJdev): evaluate validation of the found file path
//
// findKubeConfig searches the filesystem for possible locations of a KubeConfig file, which is most commonly
// located at "$HOME/.kube/config". In addition to the aforementioned path we add "$HOME/.config/gopskit/kubeconfig" as
// another possible location for the configuration file
//
// The function returns the first path which exists and (for now) does zero checking if we're actually looking at a
// possibly valid file
func findKubeConfig() (string, error) {
	for _, path := range searchPaths {
		exists := filesystem.CheckIfExists(path)
		if exists {
			return path, nil
		} else {
			continue
		}
	}

	return "", errors.New("could not find kubeconfig at either of the search paths: 1. " + searchPaths[0] + ". 2." + searchPaths[1])
}
