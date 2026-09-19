package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Client) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return c.Kube.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) ListPods(ctx context.Context, namespace string, selector labels.Selector) (*corev1.PodList, error) {
	opts := metav1.ListOptions{}
	if selector != nil && !selector.Empty() {
		opts.LabelSelector = selector.String()
	}
	return c.Kube.CoreV1().Pods(namespace).List(ctx, opts)
}

func (c *Client) ListPodsForLabels(ctx context.Context, namespace string, set map[string]string) (*corev1.PodList, error) {
	return c.ListPods(ctx, namespace, labels.SelectorFromSet(set))
}

func OwnerName(pod *corev1.Pod, kind string) string {
	for _, ref := range pod.OwnerReferences {
		if ref.Kind == kind && ref.Controller != nil && *ref.Controller {
			return ref.Name
		}
	}
	return ""
}
