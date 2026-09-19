package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpListsNewCommands(t *testing.T) {
	root := New()
	root.SetArgs([]string{"--help"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"investigate", "scan", "pods", "events", "nodes", "namespaces", "current-context", "explain", "--fail", "--quiet"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q:\n%s", want, out)
		}
	}
}

func TestInvestigateHelpHasExamples(t *testing.T) {
	root := New()
	root.SetArgs([]string{"investigate", "--help"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"--timeout", "--events", "--no-graph", "--log-tail", "--ai", "banking-dev"} {
		if !strings.Contains(out, want) {
			t.Fatalf("investigate help missing %q:\n%s", want, out)
		}
	}
}

func TestExplainIMAGEPull(t *testing.T) {
	root := New()
	root.SetArgs([]string{"explain", "IMAGE_PULL"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "IMAGE_PULL") {
		t.Fatalf("%s", buf.String())
	}
	if !strings.Contains(buf.String(), "Common causes:") {
		t.Fatalf("explain should list causes:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "Investigate with:") {
		t.Fatalf("explain should point at investigate:\n%s", buf.String())
	}
}

func TestExplainIMAGEPullAlias(t *testing.T) {
	root := New()
	root.SetArgs([]string{"explain", "image-pull"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "IMAGE_PULL") {
		t.Fatalf("%s", buf.String())
	}
}

func TestExplainUnknown(t *testing.T) {
	root := New()
	root.SetArgs([]string{"explain", "NOT_A_CODE"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestExitCode(t *testing.T) {
	if ExitCode(nil) != 0 {
		t.Fatal("nil should be 0")
	}
	if ExitCode(ErrFinding) != 2 {
		t.Fatal("finding should be 2")
	}
}

func TestVersionOmitsDev(t *testing.T) {
	root := New()
	root.SetArgs([]string{"version"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "kubecone 0.2.0") {
		t.Fatalf("missing 0.2.0: %s", out)
	}
	if strings.Contains(out, "(dev)") {
		t.Fatalf("still prints (dev): %s", out)
	}
}

func TestPodsHelp(t *testing.T) {
	root := New()
	root.SetArgs([]string{"pods", "--help"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"--unhealthy", "--all-namespaces", "banking-dev"} {
		if !strings.Contains(out, want) {
			t.Fatalf("pods help missing %q:\n%s", want, out)
		}
	}
}
