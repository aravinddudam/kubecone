package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type ScanReport struct {
	Items []evidence.ScanItem `json:"items"`
	Count int                 `json:"count"`
}

func WriteScan(out io.Writer, items []evidence.ScanItem, format string) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(ScanReport{Items: items, Count: len(items)})
	case "text", "terminal", "":
		return scanTable(out, items)
	default:
		return fmt.Errorf("unknown output format %q (use text or json)", format)
	}
}

func scanTable(out io.Writer, items []evidence.ScanItem) error {
	if len(items) == 0 {
		fmt.Fprintln(out, "No matching workloads.")
		return nil
	}
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tKIND\tNAME\tREADY\tSTATUS\tCODE")
	for _, item := range items {
		code := item.Code
		if code == "" {
			code = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Namespace, item.Kind, item.Name, item.Ready, item.Status, code)
	}
	return w.Flush()
}
