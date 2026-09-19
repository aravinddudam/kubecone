package engine

import (
	"context"
	"testing"

	"github.com/aravinddudam/kubecone/internal/evidence"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
)

func TestInvestigateRanksOOM(t *testing.T) {
	eng := NewWithCollector(func(context.Context, kubernetes.Target) (*evidence.Snapshot, error) {
		exit := int32(137)
		return &evidence.Snapshot{
			Target: evidence.Target{Kind: "Deployment", Name: "payment-api", Namespace: "shop"},
			Workload: &evidence.WorkloadFact{
				Kind: "Deployment", Name: "payment-api", Namespace: "shop", Replicas: 1,
			},
			Pods: []evidence.PodFact{{Name: "payment-api-0", Namespace: "shop", Phase: "Running"}},
			Containers: []evidence.ContainerFact{{
				Pod:                  "payment-api-0",
				Namespace:            "shop",
				Name:                 "app",
				RestartCount:         6,
				WaitingReason:        "CrashLoopBackOff",
				LastTerminatedReason: "OOMKilled",
				LastTerminatedExitCode: &exit,
				MemoryLimit:          "32Mi",
			}},
		}, nil
	}, nil)

	report, err := eng.Investigate(context.Background(), kubernetes.Target{
		Kind: "Deployment", Name: "payment-api", Namespace: "shop",
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Primary == nil || report.Primary.Code != evidence.CodeOOMKilled {
		t.Fatalf("primary=%v", report.Primary)
	}
	if len(report.Graph.Nodes) == 0 {
		t.Fatal("expected evidence graph nodes")
	}
}

func TestInvestigateHealthy(t *testing.T) {
	eng := NewWithCollector(func(context.Context, kubernetes.Target) (*evidence.Snapshot, error) {
		return &evidence.Snapshot{
			Target: evidence.Target{Kind: "Deployment", Name: "ok", Namespace: "default"},
			Pods:   []evidence.PodFact{{Name: "ok-0", Namespace: "default", Phase: "Running"}},
			Containers: []evidence.ContainerFact{{
				Pod: "ok-0", Namespace: "default", Name: "app", Ready: true, Started: true,
			}},
		}, nil
	}, nil)
	report, err := eng.Investigate(context.Background(), kubernetes.Target{Kind: "Deployment", Name: "ok", Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}
	if report.Primary.Code != evidence.CodeHealthy {
		t.Fatalf("got %s", report.Primary.Code)
	}
}
