package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func sampleReport() *evidence.Report {
	return &evidence.Report{
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
		AI:    &evidence.AINote{Provider: "none", Skipped: true, Reason: "deterministic diagnosis only"},
	}
}

func TestJSONAndTerminal(t *testing.T) {
	report := sampleReport()

	var jsonBuf bytes.Buffer
	if err := Write(&jsonBuf, report, Options{Format: "json"}); err != nil {
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
	if err := Write(&text, report, Options{Format: "text", HideAISkip: true, MaxEvents: 8}); err != nil {
		t.Fatal(err)
	}
	out := text.String()
	if !strings.Contains(out, "OOM_KILLED") {
		t.Fatalf("terminal output missing code: %s", out)
	}
	if strings.Contains(out, "skipped") {
		t.Fatalf("skipped AI should be hidden by default: %s", out)
	}
}

func TestQuietTerminal(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, sampleReport(), Options{Quiet: true, NoColor: true}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "OOM_KILLED") || !strings.Contains(out, "Next:") {
		t.Fatalf("quiet output missing diagnosis: %s", out)
	}
	if strings.Contains(out, "Evidence graph") {
		t.Fatalf("quiet output should omit graph: %s", out)
	}
}

func TestScanTable(t *testing.T) {
	var buf bytes.Buffer
	items := []evidence.ScanItem{{
		Namespace: "banking-dev",
		Kind:      "Deployment",
		Name:      "customer-service",
		Ready:     "0/1",
		Status:    "ImagePullBackOff",
		Code:      evidence.CodeImagePull,
	}}
	if err := WriteScan(&buf, items, "text"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "IMAGE_PULL") {
		t.Fatalf("scan table: %s", buf.String())
	}
}
