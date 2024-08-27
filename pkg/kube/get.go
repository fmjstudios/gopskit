package kube

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetOptions struct {
	Name       string
	Namespace  string
	GetOptions *metav1.GetOptions
}

func (c *KubeClient) GetPod(opts *GetOptions) (*corev1.Pod, error) {
	namespace := c.namespace
	if opts.Namespace != "" {
		namespace = opts.Namespace
	}

	pod, err := c.client.CoreV1().Pods(namespace).Get(context.Background(), opts.Name, *opts.GetOptions)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (c *KubeClient) GetService(opts *GetOptions) (*corev1.Service, error) {
	namespace := c.namespace
	if opts.Namespace != "" {
		namespace = opts.Namespace
	}

	svc, err := c.client.CoreV1().Services(namespace).Get(context.Background(), opts.Name, *opts.GetOptions)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func (c *KubeClient) GetSecret(opts *GetOptions) (*corev1.Secret, error) {
	namespace := c.namespace
	if opts.Namespace != "" {
		namespace = opts.Namespace
	}

	sec, err := c.client.CoreV1().Secrets(namespace).Get(context.Background(), opts.Name, *opts.GetOptions)
	if err != nil {
		return nil, err
	}

	return sec, nil
}

func (c *KubeClient) GetIngress(opts *GetOptions) (*networkingv1.Ingress, error) {
	namespace := c.namespace
	if opts.Namespace != "" {
		namespace = opts.Namespace
	}

	ing, err := c.client.NetworkingV1().Ingresses(namespace).Get(context.Background(), opts.Name, *opts.GetOptions)
	if err != nil {
		return nil, err
	}

	return ing, nil
}
