package reporter

import (
	"encoding/json"
	"io"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func JSON(out io.Writer, report *evidence.Report) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
