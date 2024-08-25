package kube

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// KubeExecutor is an Executor capable of executing remote commands within containers
type KubeExecutor int

const (
	// a built-in Executor relying on SPDY
	SPDYExecutor = iota

	// the default Executor using WebSockets
	WebSocketExecutor

	// an alias for the default Executor using WebSockets unless the feature is disabled
	FallbackExecutor
)

// Enum string representation
func (e KubeExecutor) String() string {
	return [...]string{"SPDY", "WebSocket", "Fallback"}[e]
}

// Exec executes a command within the container of a specific Pod in the current namespace
// configured for the client. If another namespace is required the ExecInNamespace helper
// provides an escape hatch
func (c *KubeClient) Exec(command, container, pod string) (string, string, error) {
	var err error
	var stdOut, stdErr bytes.Buffer

	args := strings.Fields(command)

	req := c.kcs.CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Name(pod).
		Namespace(c.namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   args,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := createExecutor(req.URL(), c.krc, SPDYExecutor)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdOut,
		Stderr: &stdErr,
		Tty:    false,
	})

	if err != nil {
		return "", "", fmt.Errorf("error in Kubernetes exec Stream: %v", err.Error())
	}

	return stdOut.String(), stdErr.String(), nil
}

// ExecInNamspace executes a command in a custom namespace instead of using the on that was initially configured
// for the KubeClient
func (c *KubeClient) ExecInNamespace(command, container, pod, namespace string) (string, string, error) {
	c.namespace = namespace
	return c.Exec(command, container, pod)
}

// crateExecutor creates a Kubernetes remote Executor for use with the Exec or ExecInNamspace client methods
// NOTE: This function is largely analogous to the implementation within the `kubectl exec` command
// ref: https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/kubectl/pkg/cmd/exec/exec.go#L138
func createExecutor(url *url.URL, config *rest.Config, executor KubeExecutor) (remotecommand.Executor, error) {
	var exec, websocketExec remotecommand.Executor = nil, nil
	var err error

	switch executor {

	case SPDYExecutor:
		exec, err = remotecommand.NewSPDYExecutor(config, "POST", url)

	case WebSocketExecutor:
		exec, err = remotecommand.NewWebSocketExecutor(config, "GET", url.String())

	case FallbackExecutor:
		// Fallback executor is default, unless feature flag is explicitly disabled.
		websocketExec, err = remotecommand.NewWebSocketExecutor(config, "GET", url.String())
		if err != nil {
			break
		}

		exec, err = remotecommand.NewFallbackExecutor(websocketExec, exec, func(err error) bool {
			return httpstream.IsUpgradeFailure(err) || httpstream.IsHTTPSProxyError(err)
		})
	}

	if err != nil {
		return nil, err
	}

	return exec, nil
}
