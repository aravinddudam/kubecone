package analyzer

import (
	"fmt"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type Probe struct{}

func (Probe) Name() string { return "probe" }

func (Probe) Analyze(snap evidence.Snapshot) []evidence.Finding {
	var findings []evidence.Finding
	seen := map[string]struct{}{}

	for _, pod := range snap.Pods {
		if strings.EqualFold(pod.Phase, "Pending") {
			continue
		}
		if isOOMPod(snap, pod.Name) || crashLooping(snap, pod.Name) || imagePulling(snap, pod.Name) {
			continue
		}
		for _, cond := range pod.Conditions {
			kind := probeKind(cond)
			if kind == "" {
				continue
			}
			key := pod.Name + "/" + kind
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			findings = append(findings, evidence.Finding{
				Code:     evidence.CodeProbeFailed,
				Severity: evidence.SevHigh,
				Title:    kind + " probe is failing",
				Summary: fmt.Sprintf(
					"Pod %s is not passing its %s probe (%s). %s",
					pod.Name, strings.ToLower(kind), cond.Reason, trim(cond.Message, 200),
				),
				Resource:       resourceName(pod.Namespace, "Pod", pod.Name),
				Recommendation: probeAdvice(kind, cond.Message),
				Confidence:     0.88,
				Attributes: map[string]string{
					"probe":  kind,
					"reason": cond.Reason,
				},
			})
		}
	}

	for _, ev := range snap.Events {
		if !strings.EqualFold(ev.Reason, "Unhealthy") {
			continue
		}
		kind := "Readiness"
		if strings.Contains(strings.ToLower(ev.Message), "liveness") {
			kind = "Liveness"
		}
		key := ev.InvolvedName + "/" + kind
		if _, ok := seen[key]; ok {
			continue
		}
		if isOOMPod(snap, ev.InvolvedName) || crashLooping(snap, ev.InvolvedName) || imagePulling(snap, ev.InvolvedName) {
			continue
		}
		seen[key] = struct{}{}
		findings = append(findings, evidence.Finding{
			Code:           evidence.CodeProbeFailed,
			Severity:       evidence.SevHigh,
			Title:          kind + " probe is failing",
			Summary:        trim(ev.Message, 240),
			Resource:       resourceName(ev.Namespace, ev.InvolvedKind, ev.InvolvedName),
			Recommendation: probeAdvice(kind, ev.Message),
			Confidence:     0.85,
		})
	}
	return findings
}

func probeKind(cond evidence.Condition) string {
	if cond.Status == "True" {
		return ""
	}
	switch cond.Type {
	case "Ready":
		if cond.Reason == "ContainersNotReady" || strings.Contains(strings.ToLower(cond.Message), "readiness") {
			return "Readiness"
		}
	case "ContainersReady":
		if strings.Contains(strings.ToLower(cond.Message), "probe") {
			return "Readiness"
		}
	}
	return ""
}

func probeAdvice(kind, message string) string {
	if strings.Contains(strings.ToLower(message), "connection refused") {
		return "The probe target is not listening yet. Fix the port, path, or give the process time to bind."
	}
	if kind == "Liveness" {
		return "A failing liveness probe restarts the container. Confirm the path/port or lengthen the probe."
	}
	return "A failing readiness probe keeps the pod out of Service endpoints. Confirm the probe path, port, and initial delay."
}

func isOOMPod(snap evidence.Snapshot, pod string) bool {
	for _, c := range snap.Containers {
		if c.Pod == pod && isOOM(c) {
			return true
		}
	}
	return false
}

func crashLooping(snap evidence.Snapshot, pod string) bool {
	for _, c := range snap.Containers {
		if c.Pod == pod && isCrashLoop(c) {
			return true
		}
	}
	return false
}

func imagePulling(snap evidence.Snapshot, pod string) bool {
	for _, c := range snap.Containers {
		if c.Pod == pod && looksLikeImagePull(c.WaitingReason, c.WaitingMessage) {
			return true
		}
	}
	return false
}
