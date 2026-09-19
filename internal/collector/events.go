package collector

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/aravinddudam/kubecone/internal/evidence"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
)

func (c *Collector) collectEvents(ctx context.Context, target kubernetes.Target, pods []corev1.Pod, snap *evidence.Snapshot) {
	list, err := c.client.ListEvents(ctx, target.Namespace)
	if err != nil {
		snap.Warnings = append(snap.Warnings, "events: "+err.Error())
		return
	}

	names := map[string]struct{}{target.Name: {}}
	for _, p := range pods {
		names[p.Name] = struct{}{}
	}
	if snap.Workload != nil {
		names[snap.Workload.Name] = struct{}{}
	}

	for _, ev := range list.Items {
		if _, ok := names[ev.InvolvedObject.Name]; !ok {
			continue
		}
		last := ev.LastTimestamp.Time
		if last.IsZero() {
			last = ev.EventTime.Time
		}
		if last.IsZero() {
			last = ev.CreationTimestamp.Time
		}
		snap.Events = append(snap.Events, evidence.EventFact{
			Name:         ev.Name,
			Namespace:    ev.Namespace,
			Type:         ev.Type,
			Reason:       ev.Reason,
			Message:      strings.TrimSpace(ev.Message),
			Count:        ev.Count,
			InvolvedKind: ev.InvolvedObject.Kind,
			InvolvedName: ev.InvolvedObject.Name,
			LastSeen:     last,
		})
	}
}
