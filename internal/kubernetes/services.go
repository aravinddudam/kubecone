package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetService(ctx context.Context, namespace, name string) (*corev1.Service, error) {
	return c.Kube.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) ListServices(ctx context.Context, namespace string) (*corev1.ServiceList, error) {
	return c.Kube.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
}

func (c *Client) ListNodes(ctx context.Context) (*corev1.NodeList, error) {
	return c.Kube.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
}

func (c *Client) GetNode(ctx context.Context, name string) (*corev1.Node, error) {
	return c.Kube.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
}
