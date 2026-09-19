package analyzer

import (
	"fmt"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type ImagePull struct{}

func (ImagePull) Name() string { return "imagepull" }

func (ImagePull) Analyze(snap evidence.Snapshot) []evidence.Finding {
	var findings []evidence.Finding
	seen := map[string]struct{}{}

	for _, c := range snap.Containers {
		if !looksLikeImagePull(c.WaitingReason, c.WaitingMessage) {
			continue
		}
		key := c.Pod + "/" + c.Name
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		msg := firstNonEmpty(c.WaitingMessage, "image could not be pulled")
		findings = append(findings, evidence.Finding{
			Code:     evidence.CodeImagePull,
			Severity: evidence.SevCritical,
			Title:    "Image cannot be pulled",
			Summary: fmt.Sprintf(
				"%s/%s cannot start because image %q failed to pull (%s). %s",
				c.Pod, c.Name, valueOr(c.Image, "unknown"), c.WaitingReason, trim(msg, 200),
			),
			Resource:       resourceName(c.Namespace, "Pod", c.Pod),
			Recommendation: imagePullAdvice(c.WaitingMessage, c.Image),
			Confidence:     0.96,
			Attributes: map[string]string{
				"container": c.Name,
				"image":     c.Image,
				"reason":    c.WaitingReason,
			},
		})
	}

	for _, ev := range snap.Events {
		if !looksLikeImagePull(ev.Reason, ev.Message) {
			continue
		}
		key := ev.InvolvedName
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		findings = append(findings, evidence.Finding{
			Code:           evidence.CodeImagePull,
			Severity:       evidence.SevCritical,
			Title:          "Image cannot be pulled",
			Summary:        trim(ev.Message, 240),
			Resource:       resourceName(ev.Namespace, ev.InvolvedKind, ev.InvolvedName),
			Recommendation: imagePullAdvice(ev.Message, ""),
			Confidence:     0.9,
		})
	}
	return findings
}

func imagePullAdvice(message, image string) string {
	m := strings.ToLower(message + " " + image)
	switch {
	case strings.Contains(m, "unauthorized") || strings.Contains(m, "denied") || strings.Contains(m, "authentication"):
		return "Create or fix an imagePullSecret and attach it to the service account or pod spec."
	case strings.Contains(m, "not found") || strings.Contains(m, "manifest unknown"):
		return "Correct the image name and tag. The registry does not have this reference."
	default:
		return "Verify the image name, tag, registry reachability, and pull credentials."
	}
}
