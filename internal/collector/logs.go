package collector

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func (c *Collector) collectLogs(ctx context.Context, pods []corev1.Pod, snap *evidence.Snapshot) {
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodPending {
			continue
		}
		containers := append([]corev1.ContainerStatus{}, pod.Status.ContainerStatuses...)
		containers = append(containers, pod.Status.InitContainerStatuses...)
		if len(containers) == 0 {
			for _, spec := range pod.Spec.Containers {
				c.appendLog(ctx, snap, pod.Namespace, pod.Name, spec.Name)
			}
			continue
		}
		for _, cs := range containers {
			if cs.RestartCount == 0 && cs.State.Waiting != nil {
				continue
			}
			c.appendLog(ctx, snap, pod.Namespace, pod.Name, cs.Name)
		}
	}
}

func (c *Collector) appendLog(ctx context.Context, snap *evidence.Snapshot, ns, pod, container string) {
	text, err := c.client.PodLogs(ctx, ns, pod, container, c.opts.LogTail)
	if err != nil {
		snap.Warnings = append(snap.Warnings, "logs "+pod+"/"+container+": "+err.Error())
		return
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if len(text) > 4000 {
		text = text[len(text)-4000:]
	}
	snap.Logs = append(snap.Logs, evidence.LogFact{
		Pod:       pod,
		Container: container,
		Namespace: ns,
		Tail:      text,
	})
}
