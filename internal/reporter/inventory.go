package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func WriteEvents(out io.Writer, items []evidence.EventItem, format string) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return writeJSON(out, map[string]any{"items": items, "count": len(items)})
	case "text", "terminal", "":
		if len(items) == 0 {
			fmt.Fprintln(out, "No matching events.")
			return nil
		}
		w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAMESPACE\tLAST\tTYPE\tREASON\tOBJECT\tMESSAGE")
		for _, ev := range items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				ev.Namespace, ev.Age, ev.Type, ev.Reason, ev.Object, clip(ev.Message, 96))
		}
		return w.Flush()
	default:
		return fmt.Errorf("unknown output format %q (use text or json)", format)
	}
}

func WriteNodes(out io.Writer, items []evidence.NodeItem, format string) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return writeJSON(out, map[string]any{"items": items, "count": len(items)})
	case "text", "terminal", "":
		if len(items) == 0 {
			fmt.Fprintln(out, "No nodes.")
			return nil
		}
		w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATUS\tROLES\tVERSION\tINTERNAL-IP")
		for _, n := range items {
			ip := n.InternalIP
			if ip == "" {
				ip = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", n.Name, n.Status, n.Roles, n.Version, ip)
		}
		return w.Flush()
	default:
		return fmt.Errorf("unknown output format %q (use text or json)", format)
	}
}

func WriteNamespaces(out io.Writer, items []evidence.NamespaceItem, format string) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return writeJSON(out, map[string]any{"items": items, "count": len(items)})
	case "text", "terminal", "":
		if len(items) == 0 {
			fmt.Fprintln(out, "No namespaces.")
			return nil
		}
		w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATUS")
		for _, ns := range items {
			fmt.Fprintf(w, "%s\t%s\n", ns.Name, ns.Status)
		}
		return w.Flush()
	default:
		return fmt.Errorf("unknown output format %q (use text or json)", format)
	}
}

func WriteContext(out io.Writer, info evidence.ContextInfo, format string) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return writeJSON(out, info)
	case "text", "terminal", "":
		fmt.Fprintf(out, "Context:   %s\n", emptyDash(info.Context))
		fmt.Fprintf(out, "Namespace: %s\n", emptyDash(info.Namespace))
		fmt.Fprintf(out, "Server:    %s\n", emptyDash(info.Server))
		return nil
	default:
		return fmt.Errorf("unknown output format %q (use text or json)", format)
	}
}

func writeJSON(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func clip(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func emptyDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
