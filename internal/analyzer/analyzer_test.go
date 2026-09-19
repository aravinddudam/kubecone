package analyzer

import (
	"strings"
	"testing"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func TestOOMPrimaryOverCrashLoop(t *testing.T) {
	code := 137
	snap := evidence.Snapshot{
		Pods: []evidence.PodFact{{Name: "payment-api-0", Namespace: "shop", Phase: "Running"}},
		Containers: []evidence.ContainerFact{{
			Pod:                  "payment-api-0",
			Namespace:            "shop",
			Name:                 "app",
			RestartCount:         7,
			WaitingReason:        "CrashLoopBackOff",
			LastTerminatedReason: "OOMKilled",
			MemoryLimit:          "512Mi",
			TerminatedExitCode:   intPtr(int32(code)),
		}},
	}
	findings := Run(snap, nil)
	if len(findings) == 0 {
		t.Fatal("expected findings")
	}
	if findings[0].Code != evidence.CodeOOMKilled {
		t.Fatalf("primary=%s want %s (%v)", findings[0].Code, evidence.CodeOOMKilled, codes(findings))
	}
	if !strings.Contains(findings[0].Recommendation, "768") && !strings.Contains(findings[0].Recommendation, "Raise memory") {
		t.Fatalf("expected memory recommendation, got %s", findings[0].Recommendation)
	}
}

func TestCrashLoop(t *testing.T) {
	exit := int32(1)
	snap := evidence.Snapshot{
		Containers: []evidence.ContainerFact{{
			Pod:                    "web-0",
			Namespace:              "default",
			Name:                   "app",
			RestartCount:           4,
			WaitingReason:          "CrashLoopBackOff",
			LastTerminatedReason:   "Error",
			LastTerminatedExitCode: &exit,
		}},
		Logs: []evidence.LogFact{{
			Pod:       "web-0",
			Container: "app",
			Tail:      "panic: missing DATABASE_URL",
		}},
	}
	findings := Run(snap, nil)
	if Primary(findings).Code != evidence.CodeCrashLoop {
		t.Fatalf("got %s", codes(findings))
	}
}

func TestImagePull(t *testing.T) {
	snap := evidence.Snapshot{
		Pods: []evidence.PodFact{{Name: "api-0", Namespace: "default", Phase: "Pending"}},
		Containers: []evidence.ContainerFact{{
			Pod:            "api-0",
			Namespace:      "default",
			Name:           "app",
			Image:          "example.invalid/missing:v1",
			WaitingReason:  "ImagePullBackOff",
			WaitingMessage: `Failed to pull image "example.invalid/missing:v1": not found`,
		}},
	}
	findings := Run(snap, nil)
	if Primary(findings).Code != evidence.CodeImagePull {
		t.Fatalf("got %s", codes(findings))
	}
}

func TestScheduling(t *testing.T) {
	snap := evidence.Snapshot{
		Pods: []evidence.PodFact{{
			Name:      "batch-0",
			Namespace: "default",
			Phase:     "Pending",
			Conditions: []evidence.Condition{{
				Type:    "PodScheduled",
				Status:  "False",
				Reason:  "Unschedulable",
				Message: `0/1 nodes are available: 1 node(s) didn't match Pod's node selector.`,
			}},
		}},
		Events: []evidence.EventFact{{
			Name:         "batch-0.1",
			Namespace:    "default",
			Type:         "Warning",
			Reason:       "FailedScheduling",
			Message:      `0/1 nodes are available: 1 node(s) didn't match Pod's node selector.`,
			InvolvedKind: "Pod",
			InvolvedName: "batch-0",
		}},
	}
	findings := Run(snap, nil)
	if Primary(findings).Code != evidence.CodeFailedScheduling {
		t.Fatalf("got %s", codes(findings))
	}
}

func TestProbe(t *testing.T) {
	snap := evidence.Snapshot{
		Pods: []evidence.PodFact{{
			Name:      "web-0",
			Namespace: "default",
			Phase:     "Running",
			Conditions: []evidence.Condition{{
				Type:    "Ready",
				Status:  "False",
				Reason:  "ContainersNotReady",
				Message: `containers with unready status: [app] readiness probe failed: HTTP probe failed with statuscode: 404`,
			}},
		}},
		Containers: []evidence.ContainerFact{{
			Pod:       "web-0",
			Namespace: "default",
			Name:      "app",
			Ready:     false,
		}},
		Events: []evidence.EventFact{{
			Reason:       "Unhealthy",
			Message:      "Readiness probe failed: HTTP probe failed with statuscode: 404",
			InvolvedKind: "Pod",
			InvolvedName: "web-0",
			Namespace:    "default",
			Type:         "Warning",
		}},
	}
	findings := Run(snap, nil)
	if Primary(findings).Code != evidence.CodeProbeFailed {
		t.Fatalf("got %s", codes(findings))
	}
}

func TestHealthyHasNoFailureCodes(t *testing.T) {
	snap := evidence.Snapshot{
		Pods: []evidence.PodFact{{Name: "ok", Namespace: "default", Phase: "Running"}},
		Containers: []evidence.ContainerFact{{
			Pod: "ok", Namespace: "default", Name: "app", Ready: true, Started: true,
		}},
	}
	if findings := Run(snap, nil); len(findings) != 0 {
		t.Fatalf("expected no findings, got %s", codes(findings))
	}
}

func codes(findings []evidence.Finding) []string {
	out := make([]string, 0, len(findings))
	for _, f := range findings {
		out = append(out, f.Code)
	}
	return out
}

func intPtr(v int32) *int32 { return &v }
