package collector

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

// collectMetrics is a v0.2 placeholder. Metrics Server is useful for a quick
// CPU/memory snapshot, but it is not a monitoring backend. Prometheus will be
// wired here later without changing the Snapshot shape.
func (c *Collector) collectMetrics(ctx context.Context, pods []corev1.Pod, snap *evidence.Snapshot) {
	_ = ctx
	_ = pods
	if len(snap.Warnings) == 0 {
		return
	}
}
