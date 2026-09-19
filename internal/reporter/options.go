package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type Options struct {
	Format     string
	Quiet      bool
	NoGraph    bool
	NoColor    bool
	HideAISkip bool
	MaxEvents  int
}

func Write(out io.Writer, report *evidence.Report, opts Options) error {
	format := strings.ToLower(strings.TrimSpace(opts.Format))
	switch format {
	case "json":
		return JSON(out, report)
	case "text", "terminal", "":
		return Terminal(out, report, opts)
	default:
		return fmt.Errorf("unknown output format %q (use text or json)", format)
	}
}
