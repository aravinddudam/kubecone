package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func Terminal(out io.Writer, report *evidence.Report) error {
	c := colorEnabled()
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

	fmt.Fprintln(out, bold("KubeCone investigation"))
	fmt.Fprintf(out, "Target:  %s/%s/%s\n", report.Target.Namespace, report.Target.Kind, report.Target.Name)
	if report.Cluster != "" {
		fmt.Fprintf(out, "Context: %s\n", report.Cluster)
	}
	fmt.Fprintf(out, "When:    %s\n", report.CollectedAt.UTC().Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintln(out)

	if report.Primary != nil {
		fmt.Fprintln(out, bold("Primary diagnosis"))
		fmt.Fprintf(out, "  [%s] %s\n", sev(report.Primary.Severity), report.Primary.Title)
		fmt.Fprintf(out, "  Code:     %s\n", report.Primary.Code)
		fmt.Fprintf(out, "  Resource: %s\n", report.Primary.Resource)
		fmt.Fprintf(out, "  %s\n", report.Primary.Summary)
		if report.Primary.Recommendation != "" {
			fmt.Fprintf(out, "  Next:     %s\n", report.Primary.Recommendation)
		}
		fmt.Fprintln(out)
	}

	if len(report.Findings) > 1 {
		fmt.Fprintln(out, bold("Related findings"))
		for i, f := range report.Findings[1:] {
			fmt.Fprintf(out, "  %d. [%s] %s (%s)\n", i+1, sev(f.Severity), f.Title, f.Code)
			fmt.Fprintf(out, "     %s\n", f.Summary)
		}
		fmt.Fprintln(out)
	}

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

	if len(report.Evidence.Events) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, bold("Recent events"))
		max := 8
		if len(report.Evidence.Events) < max {
			max = len(report.Evidence.Events)
		}
		for _, ev := range report.Evidence.Events[:max] {
			fmt.Fprintf(out, "  [%s] %s %s: %s\n", ev.Type, ev.InvolvedName, ev.Reason, trimLine(ev.Message, 120))
		}
	}

	if report.AI != nil {
		fmt.Fprintln(out)
		fmt.Fprintln(out, bold("AI"))
		if report.AI.Skipped {
			fmt.Fprintf(out, "  skipped (%s): %s\n", report.AI.Provider, report.AI.Reason)
		} else {
			fmt.Fprintf(out, "  %s\n", report.AI.Text)
		}
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
