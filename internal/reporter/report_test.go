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
		Findings: []evidence.Finding{
			{
				Code:     evidence.CodeOOMKilled,
				Severity: evidence.SevCritical,
				Title:    "Container killed because it ran out of memory",
				Summary:  "payment-api-0/app was OOMKilled",
				Resource: "shop/Pod/payment-api-0",
			},
			{
				Code:     evidence.CodeOOMKilled,
				Severity: evidence.SevCritical,
				Title:    "Container killed because it ran out of memory",
				Summary:  "payment-api-0/app was OOMKilled",
				Resource: "shop/Pod/payment-api-0",
			},
		},
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
	if strings.Contains(out, "Related findings") {
		t.Fatalf("duplicate primary should not appear as related: %s", out)
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
	if !strings.Contains(buf.String(), "KubeCone scan") {
		t.Fatalf("scan should print a header: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "kubecone investigate") {
		t.Fatalf("scan should tell the user to investigate: %s", buf.String())
	}
}

func TestScanJSONEmptyFindings(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteScanMeta(&buf, nil, "json", ScanMeta{Context: "default", Namespace: "banking-dev"}); err != nil {
		t.Fatal(err)
	}
	raw := buf.String()
	if strings.Contains(raw, `"findings": null`) {
		t.Fatalf("empty scan json must not use null findings: %s", raw)
	}
	var parsed ScanReport
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Findings == nil {
		t.Fatal("findings should be empty slice")
	}
	if parsed.Namespace != "banking-dev" || parsed.Count != 0 {
		t.Fatalf("%+v", parsed)
	}
}

func TestWriteEmptyAndContext(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteScan(&buf, nil, "text"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No matching workloads.") {
		t.Fatalf("empty scan: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "KubeCone scan") {
		t.Fatalf("empty scan must still print a header: %s", buf.String())
	}

	buf.Reset()
	if err := WritePods(&buf, nil, "text"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No matching pods.") {
		t.Fatalf("empty pods: %s", buf.String())
	}

	buf.Reset()
	if err := WriteContext(&buf, evidence.ContextInfo{Context: "k3s", Namespace: "default", Server: "https://127.0.0.1"}, "text"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "k3s") || !strings.Contains(buf.String(), "https://127.0.0.1") {
		t.Fatalf("context: %s", buf.String())
	}
}

func TestTerminalAI(t *testing.T) {
	report := sampleReport()
	report.AI = &evidence.AINote{
		Provider: "openai",
		Text:     "Likely cause:\nMissing image tag.\n\nWhat to do next:\n1. Check ECR.",
	}
	var buf bytes.Buffer
	if err := Write(&buf, report, Options{Format: "text", HideAISkip: false, MaxEvents: 0}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "AI") || !strings.Contains(out, "Missing image tag") || !strings.Contains(out, "openai") {
		t.Fatalf("ai section: %s", out)
	}
}
