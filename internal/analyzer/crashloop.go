package analyzer

import (
	"fmt"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type CrashLoop struct{}

func (CrashLoop) Name() string { return "crashloop" }

func (CrashLoop) Analyze(snap evidence.Snapshot) []evidence.Finding {
	var findings []evidence.Finding
	for _, c := range snap.Containers {
		if isOOM(c) {
			continue
		}
		if looksLikeImagePull(c.WaitingReason, c.WaitingMessage) {
			continue
		}
		if !isCrashLoop(c) {
			continue
		}
		exit := "unknown"
		if c.LastTerminatedExitCode != nil {
			exit = fmt.Sprintf("%d", *c.LastTerminatedExitCode)
		} else if c.TerminatedExitCode != nil {
			exit = fmt.Sprintf("%d", *c.TerminatedExitCode)
		}
		reason := firstNonEmpty(c.WaitingReason, c.LastTerminatedReason, c.TerminatedReason, "Error")
		msg := firstNonEmpty(c.WaitingMessage, c.LastTerminatedMessage, c.TerminatedMessage)
		logHint := relatedLog(snap, c)
		summary := fmt.Sprintf("%s/%s is crash-looping (restarts=%d, last exit=%s, reason=%s).", c.Pod, c.Name, c.RestartCount, exit, reason)
		if msg != "" {
			summary += " " + trim(msg, 180)
		}
		if logHint != "" {
			summary += " Logs: " + trim(logHint, 160)
		}
		findings = append(findings, evidence.Finding{
			Code:           evidence.CodeCrashLoop,
			Severity:       evidence.SevHigh,
			Title:          "Container is crash looping",
			Summary:        summary,
			Resource:       resourceName(c.Namespace, "Pod", c.Pod),
			Recommendation: "Inspect the process exit and recent logs. Fix the command, config, or dependency the container needs to stay running.",
			Confidence:     0.9,
			Attributes: map[string]string{
				"container":    c.Name,
				"restartCount": fmt.Sprintf("%d", c.RestartCount),
				"exitCode":     exit,
			},
		})
	}
	return findings
}

func isCrashLoop(c evidence.ContainerFact) bool {
	if strings.EqualFold(c.WaitingReason, "CrashLoopBackOff") {
		return true
	}
	return c.RestartCount >= 2 && (c.TerminatedReason != "" || c.LastTerminatedReason != "")
}

func relatedLog(snap evidence.Snapshot, c evidence.ContainerFact) string {
	for _, log := range snap.Logs {
		if log.Pod == c.Pod && (log.Container == c.Name || log.Container == "") {
			return lastLine(log.Tail)
		}
	}
	return ""
}

func lastLine(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func trim(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
