package analyzer

import (
	"fmt"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type Scheduling struct{}

func (Scheduling) Name() string { return "scheduling" }

func (Scheduling) Analyze(snap evidence.Snapshot) []evidence.Finding {
	var findings []evidence.Finding
	seen := map[string]struct{}{}

	for _, pod := range snap.Pods {
		if !unschedulable(pod) {
			continue
		}
		msg := schedulingMessage(pod, snap)
		seen[pod.Name] = struct{}{}
		findings = append(findings, evidence.Finding{
			Code:           evidence.CodeFailedScheduling,
			Severity:       evidence.SevHigh,
			Title:          "Pod cannot be scheduled",
			Summary:        fmt.Sprintf("Pod %s is Pending. %s", pod.Name, trim(msg, 240)),
			Resource:       resourceName(pod.Namespace, "Pod", pod.Name),
			Recommendation: schedulingAdvice(msg),
			Confidence:     0.92,
			Attributes: map[string]string{
				"phase":  pod.Phase,
				"reason": pod.Reason,
			},
		})
	}

	for _, ev := range snap.Events {
		if !strings.EqualFold(ev.Reason, "FailedScheduling") {
			continue
		}
		if _, ok := seen[ev.InvolvedName]; ok {
			continue
		}
		seen[ev.InvolvedName] = struct{}{}
		findings = append(findings, evidence.Finding{
			Code:           evidence.CodeFailedScheduling,
			Severity:       evidence.SevHigh,
			Title:          "Pod cannot be scheduled",
			Summary:        trim(ev.Message, 240),
			Resource:       resourceName(ev.Namespace, ev.InvolvedKind, ev.InvolvedName),
			Recommendation: schedulingAdvice(ev.Message),
			Confidence:     0.9,
		})
	}
	return findings
}

func unschedulable(pod evidence.PodFact) bool {
	if !strings.EqualFold(pod.Phase, "Pending") {
		return false
	}
	if pod.Node != "" {
		return false
	}
	for _, cond := range pod.Conditions {
		if cond.Type == "PodScheduled" && cond.Status == "False" {
			return true
		}
	}
	return pod.Reason == "Unschedulable" || pod.Message != ""
}

func schedulingMessage(pod evidence.PodFact, snap evidence.Snapshot) string {
	for _, cond := range pod.Conditions {
		if cond.Type == "PodScheduled" && cond.Message != "" {
			return cond.Message
		}
	}
	for _, ev := range snap.Events {
		if ev.InvolvedName == pod.Name && strings.EqualFold(ev.Reason, "FailedScheduling") {
			return ev.Message
		}
	}
	return firstNonEmpty(pod.Message, pod.Reason, "scheduler did not bind a node")
}

func schedulingAdvice(message string) string {
	m := strings.ToLower(message)
	switch {
	case strings.Contains(m, "node selector") || strings.Contains(m, "didn't match pod selector") || strings.Contains(m, "match node selector"):
		return "Relax the nodeSelector or affinity rules, or label a node so the pod can land."
	case strings.Contains(m, "taint") || strings.Contains(m, "toleration"):
		return "Add a matching toleration, or remove the node taint that is blocking this pod."
	case strings.Contains(m, "insufficient") || strings.Contains(m, "cpu") || strings.Contains(m, "memory"):
		return "Lower resource requests or add node capacity. The scheduler cannot fit this pod."
	case strings.Contains(m, "persistentvolumeclaim") || strings.Contains(m, "pvc"):
		return "Fix the PersistentVolumeClaim binding before the pod can schedule."
	default:
		return "Inspect FailedScheduling events: nodeSelector, taints, resources, and volumes are the usual causes."
	}
}
