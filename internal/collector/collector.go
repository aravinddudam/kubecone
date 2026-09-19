package collector

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/aravinddudam/kubecone/internal/evidence"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
)

type Options struct {
	LogTail int64
}

type Collector struct {
	client *kubernetes.Client
	opts   Options
}

func New(client *kubernetes.Client, opts Options) *Collector {
	if opts.LogTail <= 0 {
		opts.LogTail = kubernetes.DefaultLogTail
	}
	return &Collector{client: client, opts: opts}
}

func (c *Collector) Collect(ctx context.Context, target kubernetes.Target) (*evidence.Snapshot, error) {
	snap := &evidence.Snapshot{
		Target: evidence.Target{
			Kind:      target.Kind,
			Name:      target.Name,
			Namespace: target.Namespace,
		},
		Cluster:     c.client.Context,
		CollectedAt: time.Now().UTC(),
	}

	pods, err := c.collectWorkload(ctx, target, snap)
	if err != nil {
		return nil, err
	}

	snap.Pods = podFacts(pods)
	snap.Containers = containerFacts(pods)
	c.collectEvents(ctx, target, pods, snap)
	c.collectLogs(ctx, pods, snap)
	c.collectServices(ctx, target, pods, snap)
	c.collectNodes(ctx, pods, snap)
	c.collectMetrics(ctx, pods, snap)
	return snap, nil
}

func (c *Collector) collectWorkload(ctx context.Context, target kubernetes.Target, snap *evidence.Snapshot) ([]corev1.Pod, error) {
	switch target.Kind {
	case "Pod":
		pod, err := c.client.GetPod(ctx, target.Namespace, target.Name)
		if err != nil {
			return nil, fmt.Errorf("get pod: %w", err)
		}
		return []corev1.Pod{*pod}, nil
	case "Deployment":
		d, err := c.client.GetDeployment(ctx, target.Namespace, target.Name)
		if err != nil {
			return nil, fmt.Errorf("get deployment: %w", err)
		}
		snap.Workload = deploymentFact(d)
		return c.podsFor(ctx, d.Namespace, d.Spec.Selector)
	case "ReplicaSet":
		rs, err := c.client.GetReplicaSet(ctx, target.Namespace, target.Name)
		if err != nil {
			return nil, fmt.Errorf("get replicaset: %w", err)
		}
		snap.Workload = replicaSetFact(rs)
		return c.podsFor(ctx, rs.Namespace, rs.Spec.Selector)
	case "StatefulSet":
		sts, err := c.client.GetStatefulSet(ctx, target.Namespace, target.Name)
		if err != nil {
			return nil, fmt.Errorf("get statefulset: %w", err)
		}
		snap.Workload = statefulSetFact(sts)
		return c.podsFor(ctx, sts.Namespace, sts.Spec.Selector)
	case "DaemonSet":
		ds, err := c.client.GetDaemonSet(ctx, target.Namespace, target.Name)
		if err != nil {
			return nil, fmt.Errorf("get daemonset: %w", err)
		}
		snap.Workload = daemonSetFact(ds)
		return c.podsFor(ctx, ds.Namespace, ds.Spec.Selector)
	case "Service":
		svc, err := c.client.GetService(ctx, target.Namespace, target.Name)
		if err != nil {
			return nil, fmt.Errorf("get service: %w", err)
		}
		snap.Services = append(snap.Services, serviceFact(*svc))
		if len(svc.Spec.Selector) == 0 {
			snap.Warnings = append(snap.Warnings, "service has no selector")
			return nil, nil
		}
		list, err := c.client.ListPodsForLabels(ctx, svc.Namespace, svc.Spec.Selector)
		if err != nil {
			return nil, err
		}
		return list.Items, nil
	default:
		return nil, fmt.Errorf("cannot collect kind %s", target.Kind)
	}
}

func (c *Collector) podsFor(ctx context.Context, namespace string, selector *metav1.LabelSelector) ([]corev1.Pod, error) {
	sel, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, err
	}
	if sel == nil {
		sel = labels.Everything()
	}
	list, err := c.client.ListPods(ctx, namespace, sel)
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}
	return list.Items, nil
}

func deploymentFact(d *appsv1.Deployment) *evidence.WorkloadFact {
	replicas := int32(1)
	if d.Spec.Replicas != nil {
		replicas = *d.Spec.Replicas
	}
	return &evidence.WorkloadFact{
		Kind:          "Deployment",
		Name:          d.Name,
		Namespace:     d.Namespace,
		Replicas:      replicas,
		ReadyReplicas: d.Status.ReadyReplicas,
		Unavailable:   d.Status.UnavailableReplicas,
		Conditions:    workloadConditions(d.Status.Conditions),
		Selector:      metav1.FormatLabelSelector(d.Spec.Selector),
		Generation:    d.Generation,
		Observed:      d.Status.ObservedGeneration,
	}
}

func replicaSetFact(rs *appsv1.ReplicaSet) *evidence.WorkloadFact {
	replicas := int32(1)
	if rs.Spec.Replicas != nil {
		replicas = *rs.Spec.Replicas
	}
	return &evidence.WorkloadFact{
		Kind:          "ReplicaSet",
		Name:          rs.Name,
		Namespace:     rs.Namespace,
		Replicas:      replicas,
		ReadyReplicas: rs.Status.ReadyReplicas,
		Selector:      metav1.FormatLabelSelector(rs.Spec.Selector),
	}
}

func statefulSetFact(sts *appsv1.StatefulSet) *evidence.WorkloadFact {
	replicas := int32(1)
	if sts.Spec.Replicas != nil {
		replicas = *sts.Spec.Replicas
	}
	return &evidence.WorkloadFact{
		Kind:          "StatefulSet",
		Name:          sts.Name,
		Namespace:     sts.Namespace,
		Replicas:      replicas,
		ReadyReplicas: sts.Status.ReadyReplicas,
		Selector:      metav1.FormatLabelSelector(sts.Spec.Selector),
	}
}

func daemonSetFact(ds *appsv1.DaemonSet) *evidence.WorkloadFact {
	return &evidence.WorkloadFact{
		Kind:          "DaemonSet",
		Name:          ds.Name,
		Namespace:     ds.Namespace,
		Replicas:      ds.Status.DesiredNumberScheduled,
		ReadyReplicas: ds.Status.NumberReady,
		Selector:      metav1.FormatLabelSelector(ds.Spec.Selector),
	}
}

func workloadConditions(conds []appsv1.DeploymentCondition) []evidence.Condition {
	out := make([]evidence.Condition, 0, len(conds))
	for _, c := range conds {
		out = append(out, evidence.Condition{
			Type:    string(c.Type),
			Status:  string(c.Status),
			Reason:  c.Reason,
			Message: c.Message,
		})
	}
	return out
}
