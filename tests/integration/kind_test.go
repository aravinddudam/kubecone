//go:build kind

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/aravinddudam/kubecone/internal/engine"
	"github.com/aravinddudam/kubecone/internal/evidence"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
)

func TestKindCrashLoop(t *testing.T) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl not available")
	}

	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	ns := filepath.Join(root, "tests", "incidents", "namespace.yaml")
	manifest := filepath.Join(root, "tests", "incidents", "crash-loop", "manifest.yaml")

	run(t, "kubectl", "apply", "-f", ns)
	run(t, "kubectl", "apply", "-f", manifest)
	t.Cleanup(func() {
		_ = exec.Command("kubectl", "delete", "-f", manifest, "--ignore-not-found=true").Run()
	})

	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		eng, err := engine.New(engine.Options{Namespace: "kubecone-lab", Timeout: 20 * time.Second})
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		report, err := eng.Investigate(ctx, kubernetes.Target{
			Kind: "Deployment", Name: "checkout-worker", Namespace: "kubecone-lab",
		})
		cancel()
		if err == nil && report.Primary != nil && report.Primary.Code == evidence.CodeCrashLoop {
			return
		}
		time.Sleep(3 * time.Second)
	}

	t.Fatal("timed out waiting for CRASH_LOOP diagnosis")
}

func TestExpectedFilesParse(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "incidents"))
	matches, _ := filepath.Glob(filepath.Join(root, "*", "expected.json"))
	if len(matches) == 0 {
		t.Fatal("no expected.json files")
	}
	for _, path := range matches {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
	}
}

func run(t *testing.T, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, buf.String())
	}
}
