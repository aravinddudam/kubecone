package reporter

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func Write(out io.Writer, report *evidence.Report, format string) error {
	switch strings.ToLower(format) {
	case "json":
		return JSON(out, report)
	case "text", "terminal", "":
		return Terminal(out, report)
	default:
		return fmt.Errorf("unknown output format %q", format)
	}
}

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func paint(enabled bool, code, s string) string {
	if !enabled {
		return s
	}
	return code + s + "\033[0m"
}
