package analyzer

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
)

// RecommendMemory bumps a known limit by 50%, inspired by how recommendation
// engines treat OOM as a signal that the limit is too tight. Prometheus-backed
// percentile strategies can replace this later without changing the Finding shape.
func RecommendMemory(limit, request string) string {
	current := limit
	if current == "" {
		current = request
	}
	if current == "" {
		return "Set a memory request and limit. After an OOMKill with no limit, start with a limit of 256Mi and observe actual usage."
	}
	q, err := resource.ParseQuantity(current)
	if err != nil {
		return "Raise the memory limit; current value " + current + " could not be parsed."
	}
	next := resource.NewQuantity(q.Value()*3/2, resource.BinarySI)
	return fmt.Sprintf("Raise memory limit from %s to %s (50%% headroom after OOMKill).", current, next.String())
}

func looksLikeImagePull(reason, message string) bool {
	r := strings.ToLower(reason + " " + message)
	return strings.Contains(r, "imagepullbackoff") ||
		strings.Contains(r, "errimagepull") ||
		strings.Contains(r, "failed to pull image") ||
		strings.Contains(r, "not found") && strings.Contains(r, "image") ||
		strings.Contains(r, "unauthorized") && strings.Contains(r, "image")
}
