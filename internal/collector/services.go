package collector

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/aravinddudam/kubecone/internal/evidence"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
)

func (c *Collector) collectServices(ctx context.Context, target kubernetes.Target, pods []corev1.Pod, snap *evidence.Snapshot) {
	if target.Kind == "Service" {
		return
	}
	list, err := c.client.ListServices(ctx, target.Namespace)
	if err != nil {
		snap.Warnings = append(snap.Warnings, "services: "+err.Error())
		return
	}
	for _, svc := range list.Items {
		if len(svc.Spec.Selector) == 0 {
			continue
		}
		if selectsAny(svc.Spec.Selector, pods) {
			snap.Services = append(snap.Services, serviceFact(svc))
		}
	}
}

func selectsAny(selector map[string]string, pods []corev1.Pod) bool {
	for _, pod := range pods {
		if labelsMatch(selector, pod.Labels) {
			return true
		}
	}
	return false
}

func labelsMatch(selector, labels map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return len(selector) > 0
}

func serviceFact(svc corev1.Service) evidence.ServiceFact {
	return evidence.ServiceFact{
		Name:      svc.Name,
		Namespace: svc.Namespace,
		Type:      string(svc.Spec.Type),
		Selector:  svc.Spec.Selector,
	}
}

func (c *Collector) collectNodes(ctx context.Context, pods []corev1.Pod, snap *evidence.Snapshot) {
	needed := map[string]struct{}{}
	for _, pod := range pods {
		if pod.Spec.NodeName != "" {
			needed[pod.Spec.NodeName] = struct{}{}
		}
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodScheduled && cond.Status != corev1.ConditionTrue {
				list, err := c.client.ListNodes(ctx)
				if err != nil {
					snap.Warnings = append(snap.Warnings, "nodes: "+err.Error())
					return
				}
				for _, n := range list.Items {
					snap.Nodes = append(snap.Nodes, nodeFact(n))
				}
				return
			}
		}
	}
	for name := range needed {
		n, err := c.client.GetNode(ctx, name)
		if err != nil {
			snap.Warnings = append(snap.Warnings, "node "+name+": "+err.Error())
			continue
		}
		snap.Nodes = append(snap.Nodes, nodeFact(*n))
	}
}

func nodeFact(n corev1.Node) evidence.NodeFact {
	fact := evidence.NodeFact{Name: n.Name}
	for _, cond := range n.Status.Conditions {
		fact.Conditions = append(fact.Conditions, evidence.Condition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
		if cond.Type == corev1.NodeReady {
			fact.Ready = cond.Status == corev1.ConditionTrue
		}
	}
	return fact
}
