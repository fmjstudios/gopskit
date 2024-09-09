package kube

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *KubeClient) Namespaces(opts metav1.ListOptions) ([]corev1.Namespace, error) {
	ns, err := c.Client.CoreV1().Namespaces().List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return ns.Items, nil
}

func (c *KubeClient) Pods(namespace string, opts metav1.ListOptions) ([]corev1.Pod, error) {
	podL, err := c.Client.CoreV1().Pods(namespace).List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return podL.Items, nil
}

func (c *KubeClient) Services(namespace string, opts metav1.ListOptions) ([]corev1.Service, error) {
	svcL, err := c.Client.CoreV1().Services(namespace).List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return svcL.Items, nil
}

func (c *KubeClient) Secrets(namespace string, opts metav1.ListOptions) ([]corev1.Secret, error) {
	secL, err := c.Client.CoreV1().Secrets(namespace).List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return secL.Items, nil
}

func (c *KubeClient) Ingresses(namespace string, opts metav1.ListOptions) ([]networkingv1.Ingress, error) {
	ingL, err := c.Client.NetworkingV1().Ingresses(namespace).List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return ingL.Items, nil
}
