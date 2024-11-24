package kube

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) CreatePod(namespace string, pod *corev1.Pod, opts metav1.CreateOptions) error {
	_, err := c.Client.CoreV1().Pods(namespace).Create(context.Background(), pod, opts)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateDeployment(namespace string, deployment *appsv1.Deployment, opts metav1.CreateOptions) error {
	_, err := c.Client.AppsV1().Deployments(namespace).Create(context.Background(), deployment, opts)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateService(namespace string, service *corev1.Service, opts metav1.CreateOptions) error {
	_, err := c.Client.CoreV1().Services(namespace).Create(context.Background(), service, opts)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateSecret(namespace string, secret *corev1.Secret, opts metav1.CreateOptions) error {
	_, err := c.Client.CoreV1().Secrets(namespace).Create(context.Background(), secret, opts)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateIngress(namespace string, ingress *networkingv1.Ingress, opts metav1.CreateOptions) error {
	_, err := c.Client.NetworkingV1().Ingresses(namespace).Create(context.Background(), ingress, opts)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateStorageClass(storageClass *storagev1.StorageClass, opts metav1.CreateOptions) error {
	_, err := c.Client.StorageV1().StorageClasses().Create(context.Background(), storageClass, opts)
	if err != nil {
		return err
	}

	return nil
}
