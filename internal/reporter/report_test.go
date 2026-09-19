package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func TestJSONAndTerminal(t *testing.T) {
	report := &evidence.Report{
		Target:      evidence.Target{Kind: "Deployment", Name: "payment-api", Namespace: "shop"},
		CollectedAt: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC),
		Primary: &evidence.Finding{
			Code:           evidence.CodeOOMKilled,
			Severity:       evidence.SevCritical,
			Title:          "Container killed because it ran out of memory",
			Summary:        "payment-api-0/app was OOMKilled",
			Resource:       "shop/Pod/payment-api-0",
			Recommendation: "Raise memory limit from 32Mi to 48Mi",
		},
		Findings: []evidence.Finding{{
			Code:     evidence.CodeOOMKilled,
			Severity: evidence.SevCritical,
			Title:    "Container killed because it ran out of memory",
			Summary:  "payment-api-0/app was OOMKilled",
		}},
		Graph: evidence.Graph{Nodes: []evidence.Node{{Kind: "Deployment", Name: "payment-api", Status: "0/1 ready"}}},
	}

	var jsonBuf bytes.Buffer
	if err := Write(&jsonBuf, report, "json"); err != nil {
		t.Fatal(err)
	}
	var parsed evidence.Report
	if err := json.Unmarshal(jsonBuf.Bytes(), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Primary.Code != evidence.CodeOOMKilled {
		t.Fatalf("json primary=%s", parsed.Primary.Code)
	}

	var text bytes.Buffer
	if err := Write(&text, report, "text"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text.String(), "OOM_KILLED") {
		t.Fatalf("terminal output missing code: %s", text.String())
	}
}
