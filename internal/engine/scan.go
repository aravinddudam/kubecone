package engine

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aravinddudam/kubecone/internal/analyzer"
	"github.com/aravinddudam/kubecone/internal/collector"
	"github.com/aravinddudam/kubecone/internal/evidence"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
)

type ScanOptions struct {
	Namespace     string
	AllNamespaces bool
	UnhealthyOnly bool
	Kinds         []string
}

func Scan(ctx context.Context, client *kubernetes.Client, opts ScanOptions) ([]evidence.ScanItem, error) {
	ns := namespaceOrAll(client, opts.Namespace, opts.AllNamespaces)

	kinds := opts.Kinds
	if len(kinds) == 0 {
		kinds = []string{"Deployment"}
	}

	var items []evidence.ScanItem
	for _, kind := range kinds {
		switch canonicalKind(kind) {
		case "Deployment":
			list, err := client.ListDeployments(ctx, ns)
			if err != nil {
				return nil, fmt.Errorf("list deployments: %w", err)
			}
			for i := range list.Items {
				item, err := scanDeployment(ctx, client, &list.Items[i])
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			}
		case "StatefulSet":
			list, err := client.ListStatefulSets(ctx, ns)
			if err != nil {
				return nil, fmt.Errorf("list statefulsets: %w", err)
			}
			for i := range list.Items {
				item, err := scanStatefulSet(ctx, client, &list.Items[i])
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			}
		case "DaemonSet":
			list, err := client.ListDaemonSets(ctx, ns)
			if err != nil {
				return nil, fmt.Errorf("list daemonsets: %w", err)
			}
			for i := range list.Items {
				item, err := scanDaemonSet(ctx, client, &list.Items[i])
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			}
		default:
			return nil, fmt.Errorf("unsupported scan kind %q (use deployment, statefulset, daemonset, or all)", kind)
		}
	}

	return filterUnhealthy(items, opts.UnhealthyOnly), nil
}

func ScanPods(ctx context.Context, client *kubernetes.Client, opts ScanOptions) ([]evidence.ScanItem, error) {
	ns := namespaceOrAll(client, opts.Namespace, opts.AllNamespaces)

	list, err := client.ListPods(ctx, ns, nil)
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}
	items := make([]evidence.ScanItem, 0, len(list.Items))
	for i := range list.Items {
		items = append(items, scanPod(&list.Items[i]))
	}
	return filterUnhealthy(items, opts.UnhealthyOnly), nil
}

func filterUnhealthy(items []evidence.ScanItem, only bool) []evidence.ScanItem {
	if !only {
		return items
	}
	out := make([]evidence.ScanItem, 0, len(items))
	for _, item := range items {
		if item.Unhealthy() {
			out = append(out, item)
		}
	}
	return out
}

func scanPod(pod *corev1.Pod) evidence.ScanItem {
	desired := int32(len(pod.Spec.Containers))
	var ready int32
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
	}
	return diagnoseWorkload("Pod", pod.Namespace, pod.Name, ready, desired, []corev1.Pod{*pod})
}

func ParseScanKinds(raw string) []string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "deploy" || raw == "deployment" {
		return []string{"Deployment"}
	}
	if raw == "all" {
		return []string{"Deployment", "StatefulSet", "DaemonSet"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, canonicalKind(p))
	}
	return out
}

func canonicalKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "deploy", "deployment", "deployments":
		return "Deployment"
	case "sts", "statefulset", "statefulsets":
		return "StatefulSet"
	case "ds", "daemonset", "daemonsets":
		return "DaemonSet"
	case "all":
		return "all"
	default:
		return kind
	}
}

func scanDeployment(ctx context.Context, client *kubernetes.Client, d *appsv1.Deployment) (evidence.ScanItem, error) {
	replicas := int32(1)
	if d.Spec.Replicas != nil {
		replicas = *d.Spec.Replicas
	}
	pods, err := collector.PodsForDeployment(ctx, client, d)
	if err != nil {
		return evidence.ScanItem{}, err
	}
	item := diagnoseWorkload("Deployment", d.Namespace, d.Name, d.Status.ReadyReplicas, replicas, pods)
	item.Available = fmt.Sprintf("%d", d.Status.AvailableReplicas)
	return item, nil
}

func scanStatefulSet(ctx context.Context, client *kubernetes.Client, sts *appsv1.StatefulSet) (evidence.ScanItem, error) {
	replicas := int32(1)
	if sts.Spec.Replicas != nil {
		replicas = *sts.Spec.Replicas
	}
	pods, err := podsForSelector(ctx, client, sts.Namespace, sts.Spec.Selector)
	if err != nil {
		return evidence.ScanItem{}, err
	}
	return diagnoseWorkload("StatefulSet", sts.Namespace, sts.Name, sts.Status.ReadyReplicas, replicas, pods), nil
}

func scanDaemonSet(ctx context.Context, client *kubernetes.Client, ds *appsv1.DaemonSet) (evidence.ScanItem, error) {
	pods, err := podsForSelector(ctx, client, ds.Namespace, ds.Spec.Selector)
	if err != nil {
		return evidence.ScanItem{}, err
	}
	return diagnoseWorkload("DaemonSet", ds.Namespace, ds.Name, ds.Status.NumberReady, ds.Status.DesiredNumberScheduled, pods), nil
}

func podsForSelector(ctx context.Context, client *kubernetes.Client, namespace string, selector *metav1.LabelSelector) ([]corev1.Pod, error) {
	sel, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, err
	}
	list, err := client.ListPods(ctx, namespace, sel)
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}
	return list.Items, nil
}

func diagnoseWorkload(kind, namespace, name string, ready, desired int32, pods []corev1.Pod) evidence.ScanItem {
	item := evidence.ScanItem{
		Namespace: namespace,
		Kind:      kind,
		Name:      name,
		Ready:     fmt.Sprintf("%d/%d", ready, desired),
		Resource:  namespace + "/" + kind + "/" + name,
		Status:    "Ready",
		Code:      evidence.CodeHealthy,
		Title:     "Ready",
	}
	snap := evidence.Snapshot{
		Pods:       collector.PodFacts(pods),
		Containers: collector.ContainerFacts(pods),
	}
	findings := analyzer.Run(snap, nil)
	if primary := analyzer.Primary(findings); primary != nil {
		item.Code = primary.Code
		item.Title = primary.Title
		item.Severity = strings.ToUpper(primary.Severity)
		item.Summary = primary.Summary
		item.Next = primary.Recommendation
		item.Status = waitingOrPhase(pods)
		if item.Status == "" {
			item.Status = primary.Code
		}
		if primary.Attributes != nil {
			item.Container = primary.Attributes["container"]
			item.Image = primary.Attributes["image"]
			item.Cause = primary.Attributes["classifiedReason"]
			if item.Cause == "" {
				item.Cause = primary.Attributes["cause"]
			}
		}
		if item.Pod == "" {
			item.Pod = failingPodName(pods)
		}
		return item
	}
	if desired > 0 && ready < desired {
		item.Status = waitingOrPhase(pods)
		if item.Status == "" || item.Status == string(corev1.PodRunning) {
			item.Status = "NotReady"
		}
		item.Code = ""
		item.Title = "Not fully ready"
		item.Pod = failingPodName(pods)
	}
	return item
}

func failingPodName(pods []corev1.Pod) string {
	for _, pod := range pods {
		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready || (cs.State.Waiting != nil && cs.State.Waiting.Reason != "") {
				return pod.Name
			}
		}
		if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodSucceeded {
			return pod.Name
		}
	}
	if len(pods) > 0 {
		return pods[0].Name
	}
	return ""
}

func waitingOrPhase(pods []corev1.Pod) string {
	for _, pod := range pods {
		for _, cs := range pod.Status.ContainerStatuses {
			if w := cs.State.Waiting; w != nil && w.Reason != "" {
				return w.Reason
			}
		}
		if pod.Status.Reason != "" {
			return pod.Status.Reason
		}
		if pod.Status.Phase != "" && pod.Status.Phase != corev1.PodRunning {
			return string(pod.Status.Phase)
		}
	}
	if len(pods) > 0 {
		return string(pods[0].Status.Phase)
	}
	return "NoPods"
}
