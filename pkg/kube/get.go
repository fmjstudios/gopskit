package kube

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) Namespaces(opts metav1.ListOptions) ([]corev1.Namespace, error) {
	ns, err := c.Client.CoreV1().Namespaces().List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return ns.Items, nil
}

func (c *Client) Pods(namespace string, opts metav1.ListOptions) ([]corev1.Pod, error) {
	podL, err := c.Client.CoreV1().Pods(namespace).List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return podL.Items, nil
}

func (c *Client) Service(namespace, name string, opts metav1.GetOptions) (*corev1.Service, error) {
	svc, err := c.Client.CoreV1().Services(namespace).Get(context.Background(), name, opts)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func (c *Client) Services(namespace string, opts metav1.ListOptions) ([]corev1.Service, error) {
	svcL, err := c.Client.CoreV1().Services(namespace).List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return svcL.Items, nil
}

func (c *Client) ConfigMap(namespace, name string, opts metav1.GetOptions) (*corev1.ConfigMap, error) {
	conf, err := c.Client.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, opts)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func (c *Client) ConfigMaps(namespace string, opts metav1.ListOptions) ([]corev1.ConfigMap, error) {
	confL, err := c.Client.CoreV1().ConfigMaps(namespace).List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return confL.Items, nil
}

func (c *Client) Secret(namespace string, name string, opts metav1.GetOptions) (*corev1.Secret, error) {
	sec, err := c.Client.CoreV1().Secrets(namespace).Get(context.Background(), name, opts)
	if err != nil {
		return nil, err
	}

	return sec, nil
}

func (c *Client) Secrets(namespace string, opts metav1.ListOptions) ([]corev1.Secret, error) {
	secL, err := c.Client.CoreV1().Secrets(namespace).List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return secL.Items, nil
}

func (c *Client) Ingresses(namespace string, opts metav1.ListOptions) ([]networkingv1.Ingress, error) {
	ingL, err := c.Client.NetworkingV1().Ingresses(namespace).List(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return ingL.Items, nil
}

func (c *Client) StorageClass(name string, opts metav1.GetOptions) (*storagev1.StorageClass, error) {
	storC, err := c.Client.StorageV1().StorageClasses().Get(context.Background(), name, opts)
	if err != nil {
		return nil, err
	}

	return storC, nil
}
