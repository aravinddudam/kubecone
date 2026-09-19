package ai

import (
	"fmt"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

const (
	maxEventsInPrompt = 8
	maxLogsInPrompt   = 2
	maxLogChars       = 800
	maxEventChars     = 220
)

func systemPrompt() string {
	return strings.TrimSpace(`You are a Kubernetes incident assistant for KubeCone.
KubeCone already produced a deterministic diagnosis from cluster evidence. Treat that diagnosis as the source of truth.
Do not invent cluster facts. Do not recommend imagePullSecret unless the evidence shows unauthorized/denied.
If the diagnosis is IMAGE_PULL, stay consistent with the classified cause (missing tag, auth, network, DNS, or unknown).
Write for an on-call engineer: short, specific, actionable.

Reply in this shape:
Likely cause:
<one or two sentences>

What to do next:
1. ...
2. ...
3. ...

Why this fits the evidence:
<one short paragraph>`)
}

func BuildPrompt(report *evidence.Report) string {
	if report == nil {
		return "No report."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Target: %s/%s/%s\n", report.Target.Namespace, report.Target.Kind, report.Target.Name)
	if report.Cluster != "" {
		fmt.Fprintf(&b, "Context: %s\n", report.Cluster)
	}
	if w := report.Evidence.Workload; w != nil {
		fmt.Fprintf(&b, "Workload: %s ready=%d/%d available=%d unavailable=%d\n",
			w.Name, w.ReadyReplicas, w.Replicas, w.Available, w.Unavailable)
	}
	if report.Primary != nil {
		fmt.Fprintf(&b, "\nDeterministic primary diagnosis:\n")
		fmt.Fprintf(&b, "  Code: %s\n", report.Primary.Code)
		fmt.Fprintf(&b, "  Title: %s\n", report.Primary.Title)
		fmt.Fprintf(&b, "  Summary: %s\n", report.Primary.Summary)
		if report.Primary.Recommendation != "" {
			fmt.Fprintf(&b, "  Rule next step: %s\n", report.Primary.Recommendation)
		}
		for k, v := range report.Primary.Attributes {
			if v == "" {
				continue
			}
			fmt.Fprintf(&b, "  %s: %s\n", k, v)
		}
	}
	if len(report.Findings) > 1 {
		fmt.Fprintln(&b, "\nOther findings:")
		for i, f := range report.Findings {
			if report.Primary != nil && f.Code == report.Primary.Code && f.Title == report.Primary.Title {
				continue
			}
			fmt.Fprintf(&b, "  %d. %s (%s): %s\n", i+1, f.Title, f.Code, f.Summary)
		}
	}
	if len(report.Evidence.Containers) > 0 {
		fmt.Fprintln(&b, "\nContainers:")
		for _, c := range report.Evidence.Containers {
			fmt.Fprintf(&b, "  %s/%s image=%s ready=%v waiting=%s restarts=%d\n",
				c.Pod, c.Name, c.Image, c.Ready, firstNonEmpty(c.WaitingReason, "-"), c.RestartCount)
			if c.WaitingMessage != "" {
				fmt.Fprintf(&b, "    waitingMessage: %s\n", clip(c.WaitingMessage, 300))
			}
		}
	}
	if len(report.Evidence.Events) > 0 {
		fmt.Fprintln(&b, "\nRecent events:")
		n := len(report.Evidence.Events)
		if n > maxEventsInPrompt {
			n = maxEventsInPrompt
		}
		for _, ev := range report.Evidence.Events[:n] {
			fmt.Fprintf(&b, "  [%s] %s %s: %s\n", ev.Type, ev.InvolvedName, ev.Reason, clip(ev.Message, maxEventChars))
		}
	}
	if len(report.Evidence.Logs) > 0 {
		fmt.Fprintln(&b, "\nLog tails (truncated):")
		n := len(report.Evidence.Logs)
		if n > maxLogsInPrompt {
			n = maxLogsInPrompt
		}
		for _, lg := range report.Evidence.Logs[:n] {
			fmt.Fprintf(&b, "  %s/%s:\n%s\n", lg.Pod, lg.Container, clip(lg.Tail, maxLogChars))
		}
	}
	b.WriteString("\nExplain the diagnosis and give concrete next commands/checks for this cluster.")
	return b.String()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func clip(s string, n int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
