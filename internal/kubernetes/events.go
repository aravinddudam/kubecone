package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) ListEvents(ctx context.Context, namespace string) (*corev1.EventList, error) {
	return c.Kube.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
}

func (c *Client) ListEventsForObject(ctx context.Context, namespace, kind, name string) (*corev1.EventList, error) {
	field := "involvedObject.name=" + name
	if kind != "" {
		field += ",involvedObject.kind=" + kind
	}
	return c.Kube.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{FieldSelector: field})
}
