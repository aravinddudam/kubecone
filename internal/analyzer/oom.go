package analyzer

import (
	"fmt"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type OOM struct{}

func (OOM) Name() string { return "oom" }

func (OOM) Analyze(snap evidence.Snapshot) []evidence.Finding {
	var findings []evidence.Finding
	for _, c := range snap.Containers {
		if !isOOM(c) {
			continue
		}
		findings = append(findings, evidence.Finding{
			Code:     evidence.CodeOOMKilled,
			Severity: evidence.SevCritical,
			Title:    "Container killed because it ran out of memory",
			Summary: fmt.Sprintf(
				"%s/%s was OOMKilled (restarts=%d, memory limit=%s).",
				c.Pod, c.Name, c.RestartCount, valueOr(c.MemoryLimit, "unset"),
			),
			Resource:       resourceName(c.Namespace, "Pod", c.Pod),
			Recommendation: RecommendMemory(c.MemoryLimit, c.MemoryRequest),
			Confidence:     0.95,
			Attributes: map[string]string{
				"container":    c.Name,
				"memoryLimit":  c.MemoryLimit,
				"restartCount": fmt.Sprintf("%d", c.RestartCount),
			},
		})
	}
	return findings
}

func isOOM(c evidence.ContainerFact) bool {
	return c.TerminatedReason == "OOMKilled" || c.LastTerminatedReason == "OOMKilled"
}

func valueOr(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
