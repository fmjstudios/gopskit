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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwardOptions struct {
	PodName      string
	Container    string
	Namespace    string
	Address      []string
	Ports        []string
	ReadyChannel chan struct{}
	StopChannel  chan struct{}
}

type PortForwarder interface {
	ForwardPorts(method string, restConfig *rest.Config, url *url.URL, opts PortForwardOptions) error
}

type DefaultPortForwarder struct{}

func (*DefaultPortForwarder) ForwardPorts(method string, restConfig *rest.Config, url *url.URL, opts PortForwardOptions) error {
	var err error
	var stdOut, stdErr bytes.Buffer

	dialer, err := createDialer(method, url, restConfig)
	if err != nil {
		return err
	}

	pf, err := portforward.NewOnAddresses(dialer, opts.Address, opts.Ports, opts.StopChannel, opts.StopChannel, &stdOut, &stdErr)
	if err != nil {
		return err
	}

	return pf.ForwardPorts()
}

func (c *KubeClient) PortForward(ctx context.Context, opts PortForwardOptions) error {
	if opts.Namespace != "" {
		c.namespace = opts.Namespace
	}

	podC := c.Client.CoreV1().Pods(c.namespace)

	pod, err := podC.Get(ctx, opts.PodName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("uanble to forward ports to a pod that isn't running. Current status: %v", pod.Status.Phase)
	}

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

		if opts.StopChannel != nil {
			close(opts.StopChannel)
		}
	}()

	req := c.Client.RESTClient().Post().
		Resource("pods").
		Namespace(c.namespace).
		Name(opts.PodName).SubResource("portforward")

	return c.PortForwarder.ForwardPorts("POST", c.Config, req.URL(), opts)
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
