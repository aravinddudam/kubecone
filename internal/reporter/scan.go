package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type ScanReport struct {
	Context   string              `json:"context"`
	Namespace string              `json:"namespace"`
	Findings  []evidence.ScanItem `json:"findings"`
	Count     int                 `json:"count"`
}

type ScanMeta struct {
	Title         string
	Context       string
	Namespace     string
	UnhealthyOnly bool
	EmptyHint     string
}

func WriteScan(out io.Writer, items []evidence.ScanItem, format string) error {
	return WriteScanMeta(out, items, format, ScanMeta{
		Title:     "KubeCone scan",
		EmptyHint: "No matching workloads.",
	})
}

func WritePods(out io.Writer, items []evidence.ScanItem, format string) error {
	return WriteScanMeta(out, items, format, ScanMeta{
		Title:     "KubeCone pods",
		EmptyHint: "No matching pods.",
	})
}

func WriteScanMeta(out io.Writer, items []evidence.ScanItem, format string, meta ScanMeta) error {
	if items == nil {
		items = []evidence.ScanItem{}
	}
	if meta.Title == "" {
		meta.Title = "KubeCone scan"
	}
	if meta.EmptyHint == "" {
		meta.EmptyHint = "No matching workloads."
	}
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		ns := meta.Namespace
		if ns == "" && len(items) > 0 {
			ns = items[0].Namespace
		}
		return writeJSON(out, ScanReport{
			Context:   meta.Context,
			Namespace: ns,
			Findings:  items,
			Count:     len(items),
		})
	case "text", "terminal", "":
		return scanText(out, items, meta)
	default:
		return fmt.Errorf("unknown output format %q (use text or json)", format)
	}
}

func scanText(out io.Writer, items []evidence.ScanItem, meta ScanMeta) error {
	unhealthy := 0
	for _, item := range items {
		if item.Unhealthy() {
			unhealthy++
		}
	}

	fmt.Fprintln(out, meta.Title)
	if meta.Namespace != "" {
		fmt.Fprintf(out, "Namespace: %s\n", meta.Namespace)
	}
	if meta.Context != "" {
		fmt.Fprintf(out, "Context:   %s\n", meta.Context)
	}
	fmt.Fprintf(out, "Unhealthy: %d\n", unhealthy)
	fmt.Fprintln(out)

	if len(items) == 0 {
		fmt.Fprintln(out, meta.EmptyHint)
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Next:")
		ns := meta.Namespace
		if ns == "" || ns == "(all)" {
			ns = "<namespace>"
		}
		fmt.Fprintf(out, "  kubecone pods -n %s --unhealthy\n", ns)
		fmt.Fprintf(out, "  kubecone scan -n %s --kind all\n", ns)
		return nil
	}

	wroteCard := false
	for _, item := range items {
		if !item.Unhealthy() {
			continue
		}
		writeScanCard(out, item)
		wroteCard = true
	}
	if wroteCard {
		return nil
	}

	fmt.Fprintln(out, "NAMESPACE\tKIND\tNAME\tREADY\tSTATUS\tCODE")
	for _, item := range items {
		code := item.Code
		if code == "" {
			code = "-"
		}
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Namespace, item.Kind, item.Name, item.Ready, item.Status, code)
	}
	return nil
}

func writeScanCard(out io.Writer, item evidence.ScanItem) {
	sev := item.Severity
	if sev == "" {
		sev = "CRITICAL"
	}
	fmt.Fprintf(out, "[%s] %s/%s\n", sev, item.Kind, item.Name)
	fmt.Fprintf(out, "  Ready: %s\n", item.Ready)
	if item.Code != "" {
		fmt.Fprintf(out, "  Reason: %s\n", item.Code)
	} else {
		fmt.Fprintf(out, "  Status: %s\n", item.Status)
	}
	if item.Cause != "" {
		fmt.Fprintf(out, "  Cause: %s\n", item.Cause)
	}
	if item.Pod != "" {
		fmt.Fprintf(out, "  Pod: %s\n", item.Pod)
	}
	if item.Container != "" {
		fmt.Fprintf(out, "  Container: %s\n", item.Container)
	}
	if item.Image != "" {
		fmt.Fprintf(out, "  Image: %s\n", item.Image)
	}
	fmt.Fprintln(out, "  Run:")
	fmt.Fprintf(out, "    kubecone investigate %s/%s -n %s\n", strings.ToLower(item.Kind), item.Name, item.Namespace)
	if item.Code != "" && item.Code != evidence.CodeHealthy {
		fmt.Fprintf(out, "    kubecone explain %s\n", item.Code)
	}
	fmt.Fprintln(out)
}
