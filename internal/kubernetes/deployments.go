package kubernetes

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return c.Kube.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) ListReplicaSetsForDeployment(ctx context.Context, d *appsv1.Deployment) (*appsv1.ReplicaSetList, error) {
	return c.Kube.AppsV1().ReplicaSets(d.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(d.Spec.Selector),
	})
}

func (c *Client) GetReplicaSet(ctx context.Context, namespace, name string) (*appsv1.ReplicaSet, error) {
	return c.Kube.AppsV1().ReplicaSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) GetStatefulSet(ctx context.Context, namespace, name string) (*appsv1.StatefulSet, error) {
	return c.Kube.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) GetDaemonSet(ctx context.Context, namespace, name string) (*appsv1.DaemonSet, error) {
	return c.Kube.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	return c.Kube.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
}

func (c *Client) ListStatefulSets(ctx context.Context, namespace string) (*appsv1.StatefulSetList, error) {
	return c.Kube.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
}

func (c *Client) ListDaemonSets(ctx context.Context, namespace string) (*appsv1.DaemonSetList, error) {
	return c.Kube.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
}
