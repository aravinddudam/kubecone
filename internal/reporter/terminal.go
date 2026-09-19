package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func Terminal(out io.Writer, report *evidence.Report, opts Options) error {
	c := colorEnabled(opts.NoColor)
	bold := func(s string) string { return paint(c, "\033[1m", s) }
	dim := func(s string) string { return paint(c, "\033[2m", s) }
	sev := func(s string) string {
		switch s {
		case evidence.SevCritical:
			return paint(c, "\033[31m", strings.ToUpper(s))
		case evidence.SevHigh:
			return paint(c, "\033[33m", strings.ToUpper(s))
		case evidence.SevInfo:
			return paint(c, "\033[32m", strings.ToUpper(s))
		default:
			return strings.ToUpper(s)
		}
	}

	if opts.Quiet {
		return terminalQuiet(out, report, sev)
	}

	fmt.Fprintln(out, bold("KubeCone investigation"))
	fmt.Fprintf(out, "Target:  %s/%s/%s\n", report.Target.Namespace, report.Target.Kind, report.Target.Name)
	if report.Cluster != "" {
		fmt.Fprintf(out, "Context: %s\n", report.Cluster)
	}
	fmt.Fprintf(out, "When:    %s\n", report.CollectedAt.UTC().Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintln(out)

	writeState(out, report, bold, dim)

	if report.Primary != nil {
		fmt.Fprintln(out, bold("Primary diagnosis"))
		fmt.Fprintf(out, "  [%s] %s\n", sev(report.Primary.Severity), report.Primary.Title)
		fmt.Fprintf(out, "  Code:     %s\n", report.Primary.Code)
		fmt.Fprintf(out, "  Resource: %s\n", report.Primary.Resource)
		if img := attr(report.Primary, "image"); img != "" {
			fmt.Fprintf(out, "  Image:    %s\n", img)
		}
		if reason := attr(report.Primary, "classifiedReason"); reason != "" {
			fmt.Fprintf(out, "  Reason:   %s\n", reason)
		}
		fmt.Fprintf(out, "  %s\n", report.Primary.Summary)
		if report.Primary.Recommendation != "" {
			fmt.Fprintf(out, "  Next:     %s\n", report.Primary.Recommendation)
		}
		fmt.Fprintln(out)
	}

	related := relatedFindings(report)
	if len(related) > 0 {
		fmt.Fprintln(out, bold("Related findings"))
		for i, f := range related {
			fmt.Fprintf(out, "  %d. [%s] %s (%s)\n", i+1, sev(f.Severity), f.Title, f.Code)
			fmt.Fprintf(out, "     %s\n", f.Summary)
		}
		fmt.Fprintln(out)
	}

	if !opts.NoGraph {
		fmt.Fprintln(out, bold("Evidence graph"))
		fmt.Fprintf(out, "  %d nodes, %d edges\n", len(report.Graph.Nodes), len(report.Graph.Edges))
		for _, n := range report.Graph.Nodes {
			if n.Kind == evidence.KindEvent || n.Kind == evidence.KindLog {
				continue
			}
			status := n.Status
			if status == "" {
				status = n.Kind
			}
			fmt.Fprintf(out, "  - %-12s %-40s %s\n", n.Kind, n.Name, dim(status))
		}
	}

	max := opts.MaxEvents
	if max < 0 {
		max = 8
	}
	if max > 0 && len(report.Evidence.Events) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, bold("Recent events"))
		if len(report.Evidence.Events) < max {
			max = len(report.Evidence.Events)
		}
		for _, ev := range report.Evidence.Events[:max] {
			fmt.Fprintf(out, "  [%s] %s %s: %s\n", ev.Type, ev.InvolvedName, ev.Reason, trimLine(ev.Message, 120))
		}
	}

	if report.AI != nil && !(opts.HideAISkip && report.AI.Skipped) {
		fmt.Fprintln(out)
		fmt.Fprintln(out, bold("AI"))
		if report.AI.Skipped {
			fmt.Fprintf(out, "  skipped (%s): %s\n", report.AI.Provider, report.AI.Reason)
		} else {
			fmt.Fprintf(out, "  provider: %s\n", report.AI.Provider)
			for _, line := range strings.Split(report.AI.Text, "\n") {
				fmt.Fprintf(out, "  %s\n", line)
			}
		}
	}
	return nil
}

func writeState(out io.Writer, report *evidence.Report, bold, dim func(string) string) {
	w := report.Evidence.Workload
	if w == nil && len(report.Evidence.Pods) == 0 {
		return
	}
	fmt.Fprintln(out, bold("State"))
	if w != nil {
		fmt.Fprintf(out, "  %s: %s\n", w.Kind, w.Name)
		fmt.Fprintf(out, "  Ready: %d/%d   Available: %d   Unavailable: %d\n", w.ReadyReplicas, w.Replicas, w.Available, w.Unavailable)
	}
	for _, pod := range report.Evidence.Pods {
		ready, total := 0, 0
		for _, c := range report.Evidence.Containers {
			if c.Pod != pod.Name {
				continue
			}
			total++
			if c.Ready {
				ready++
			}
		}
		fmt.Fprintf(out, "  Pod phase: %s\n", dim(valueOr(pod.Phase, "Unknown")))
		fmt.Fprintf(out, "  Container readiness: %d/%d\n", ready, total)
		for _, c := range report.Evidence.Containers {
			if c.Pod != pod.Name {
				continue
			}
			fmt.Fprintf(out, "  Container state: %s %s\n", c.Name, dim(containerState(c)))
			if c.Image != "" {
				fmt.Fprintf(out, "  Image: %s\n", c.Image)
			}
		}
		break
	}
	fmt.Fprintln(out)
}

func containerState(c evidence.ContainerFact) string {
	if c.WaitingReason != "" {
		return "Waiting (" + c.WaitingReason + ")"
	}
	if c.TerminatedReason != "" {
		return "Terminated (" + c.TerminatedReason + ")"
	}
	if c.Ready {
		return "Ready"
	}
	return "Running (not ready)"
}

func relatedFindings(report *evidence.Report) []evidence.Finding {
	if report.Primary == nil || len(report.Findings) == 0 {
		return nil
	}
	var out []evidence.Finding
	for _, f := range report.Findings {
		if f.Code == report.Primary.Code && (f.Resource == report.Primary.Resource || f.Title == report.Primary.Title) {
			continue
		}
		out = append(out, f)
	}
	return out
}

func attr(f *evidence.Finding, key string) string {
	if f == nil || f.Attributes == nil {
		return ""
	}
	return f.Attributes[key]
}

func valueOr(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func terminalQuiet(out io.Writer, report *evidence.Report, sev func(string) string) error {
	if report.Primary == nil {
		fmt.Fprintln(out, "no diagnosis")
		return nil
	}
	fmt.Fprintf(out, "[%s] %s  %s\n", sev(report.Primary.Severity), report.Primary.Code, report.Primary.Resource)
	fmt.Fprintf(out, "%s\n", report.Primary.Summary)
	if report.Primary.Recommendation != "" {
		fmt.Fprintf(out, "Next: %s\n", report.Primary.Recommendation)
	}
	if report.AI != nil && !report.AI.Skipped && strings.TrimSpace(report.AI.Text) != "" {
		fmt.Fprintln(out)
		fmt.Fprintln(out, report.AI.Text)
	}
	return nil
}

func trimLine(s string, n int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
