package kube

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type ApplyOptions struct {
	Name          string
	GetOptions    *metav1.GetOptions
	CreateOptions *metav1.CreateOptions
	UpdateOptions *metav1.UpdateOptions
}

func (c *KubeClient) Apply(schema schema.GroupVersionResource, resource *unstructured.Unstructured, opts *ApplyOptions) error {
	var result *unstructured.Unstructured
	var err error

	dc, err := dynamic.NewForConfig(c.config)
	if err != nil {
		panic(err)
	}

	result, err = dc.Resource(schema).Get(context.Background(), opts.Name, *opts.GetOptions)
	if err != nil {
		return err
	}

	exists := result.Object["metadata"] != nil
	if !exists {
		result, err = dc.Resource(schema).Create(context.Background(), resource, *opts.CreateOptions)
		if err != nil {
			return err
		}

		fmt.Printf(`Created Kubernetes Resource: %v named %v\n`, result.GetKind(), result.GetName())
		return nil
	}

	result, err = dc.Resource(schema).Update(context.Background(), resource, *opts.UpdateOptions)
	if err != nil {
		return err
	}

	fmt.Printf(`Updated Kubernetes Resource: %v named %v\n`, result.GetKind(), result.GetName())
	return nil
}
