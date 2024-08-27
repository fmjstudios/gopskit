package kube

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateOptions struct {
	Name          string
	Namespace     string
	CreateOptions *metav1.CreateOptions
}

func (c *KubeClient) CreatePod(pod *corev1.Pod, opts *CreateOptions) error {
	namespace := c.namespace
	if opts.Namespace != "" {
		namespace = opts.Namespace
	}

	_, err := c.client.CoreV1().Pods(namespace).Create(context.Background(), pod, *opts.CreateOptions)
	if err != nil {
		return err
	}

	return nil
}

func (c *KubeClient) CreateDeployment(deployment *appsv1.Deployment, opts *CreateOptions) error {
	namespace := c.namespace
	if opts.Namespace != "" {
		namespace = opts.Namespace
	}

	_, err := c.client.AppsV1().Deployments(namespace).Create(context.Background(), deployment, *opts.CreateOptions)
	if err != nil {
		return err
	}

	return nil
}

func (c *KubeClient) CreateService(service *corev1.Service, opts *CreateOptions) error {
	namespace := c.namespace
	if opts.Namespace != "" {
		namespace = opts.Namespace
	}

	_, err := c.client.CoreV1().Services(namespace).Create(context.Background(), service, *opts.CreateOptions)
	if err != nil {
		return err
	}

	return nil
}

func (c *KubeClient) CreateSecret(secret *corev1.Secret, opts *CreateOptions) error {
	namespace := c.namespace
	if opts.Namespace != "" {
		namespace = opts.Namespace
	}

	_, err := c.client.CoreV1().Secrets(namespace).Create(context.Background(), secret, *opts.CreateOptions)
	if err != nil {
		return err
	}

	return nil
}

func (c *KubeClient) CreateIngress(ingress *networkingv1.Ingress, opts *CreateOptions) error {
	namespace := c.namespace
	if opts.Namespace != "" {
		namespace = opts.Namespace
	}

	_, err := c.client.NetworkingV1().Ingresses(namespace).Create(context.Background(), ingress, *opts.CreateOptions)
	if err != nil {
		return err
	}

	return nil
}
