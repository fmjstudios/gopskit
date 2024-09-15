package kube

import (
	"errors"
	"path/filepath"

	"github.com/fmjstudios/gopskit/pkg/filesystem"
	"github.com/fmjstudios/gopskit/pkg/platform"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	DefaultNamespace = "default"
	DefaultLocalPort = "7150"
)

var (
	searchPaths []string = []string{
		filepath.Join(platform.Current().Home(), ".kube", "config"),
		filepath.Join(platform.Current().ConfigDir(), "gopskit", "kubeconfig"),
	}
)

type KubeClient struct {
	// ConfigPath is the configuration file path for which the current client(-set) was created
	ConfigPath string

	// Executor is the connection between local and remote IO streams
	Executor RemoteExecutor

	// PortForwarder is the implementation for the port-forward functionality
	PortForwarder PortForwarder

	// Namespace is the Kubernetes namespace the client is configured to access
	namespace string

	// Config is the rest.Config for which the client was built
	Config *rest.Config

	// Client is the embedded Kubernetes ClientSet
	Client *kubernetes.Clientset

	// flags are the Kubernetes-specific flags which will be injected into the CLI
	Flags *genericclioptions.ConfigFlags
}

// Opt represents a configuration option for the KubeClient
type Opt func(c *KubeClient)

// NewClient instantiates a new KubeClient with a configured and embedded Clientset
func NewClient(opts ...Opt) (*KubeClient, error) {
	var err error

	kc := &KubeClient{
		Executor:      &DefaultRemoteExecutor{},
		PortForwarder: &DefaultPortForwarder{},
	}

	for _, opt := range opts {
		opt(kc)
	}

	kc.Flags = genericclioptions.NewConfigFlags(true)

	// if WithConfigPath wasn't in the opts
	if kc.ConfigPath == "" {
		kc.ConfigPath, err = findKubeConfig()
		if err != nil {
			return nil, err
		}
	}

	if kc.namespace == "" {
		kc.namespace = DefaultNamespace
	}

	kc.Config, err = clientcmd.BuildConfigFromFlags("", kc.ConfigPath)
	if err != nil {
		return nil, err
	}

	kc.Client, err = kubernetes.NewForConfig(kc.Config)
	if err != nil {
		return nil, err
	}

	return kc, nil
}

// WithConfigPath configures the KubeClient with a predetermined path for the 'kubeconfig' file
// This avoids the usual searches done within findKubeConfig
func WithConfigPath(path string) func(c *KubeClient) {
	return func(c *KubeClient) {
		c.ConfigPath = path
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
