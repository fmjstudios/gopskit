package kube

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

//type PortForwardOptions struct {
//	PodName   string
//	Container string
//	Namespace string
//	Address   []string
//	Ports     []string
//}

type PortForwarder interface {
	ForwardPorts(method string, restConfig *rest.Config, url *url.URL, ports []string, stopChan,
		readyChan chan struct{}) error
}

type DefaultPortForwarder struct{}

func (*DefaultPortForwarder) ForwardPorts(method string, restConfig *rest.Config, url *url.URL, ports []string, stopChan,
	readyChan chan struct{}) error {
	var err error
	var stdOut, stdErr bytes.Buffer

	dialer, err := createDialer(method, url, restConfig)
	if err != nil {
		return err
	}

	pf, err := portforward.NewOnAddresses(
		dialer,
		[]string{"localhost"},
		ports,
		stopChan,
		readyChan,
		&stdOut,
		&stdErr,
	)
	if err != nil {
		return err
	}

	return pf.ForwardPorts()
}

// PortForward port-forwards a remote port of a Kubernetes container to the local machine
func (c *KubeClient) PortForward(ctx context.Context, pod corev1.Pod) error {
	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("uanble to forward ports to a pod that isn't running. Current status: %v", pod.Status.Phase)
	}

	// create control channels
	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	returnCtx, returnCtxCancel := context.WithCancel(ctx)
	defer returnCtxCancel()

	go func() {
		select {
		case <-signals:
		case <-returnCtx.Done():
		}

		if stopChan != nil {
			close(stopChan)
		}
	}()

	req := c.Client.CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Namespace(pod.Namespace).
		Name(pod.Name).
		SubResource("portforward")

	return c.PortForwarder.ForwardPorts(
		"POST",
		c.Config,
		req.URL(),
		BuildDefaultPortMap(pod.Spec.Containers[0].Ports[0].ContainerPort),
		stopChan,
		readyChan,
	)
}

func createDialer(method string, url *url.URL, restConfig *rest.Config) (httpstream.Dialer, error) {
	var dialer httpstream.Dialer
	var err error

	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return nil, err
	}

	spdyDialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	dialer, err = portforward.NewSPDYOverWebsocketDialer(url, restConfig)
	if err != nil {
		return nil, err
	}

	dialer = portforward.NewFallbackDialer(dialer, spdyDialer, func(err error) bool {
		return httpstream.IsUpgradeFailure(err) || httpstream.IsHTTPSProxyError(err)
	})

	return dialer, nil
}

// only require remoteport as input
func BuildDefaultPortMap(remotePort int32) []string {
	return []string{fmt.Sprintf("%s:%d", DefaultLocalPort, remotePort)}
}
