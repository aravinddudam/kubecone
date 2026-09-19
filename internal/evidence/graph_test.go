package evidence

import (
	"strings"
	"testing"
)

func TestBuildGraphLinksWorkloadToPod(t *testing.T) {
	snap := Snapshot{
		Workload: &WorkloadFact{Kind: KindDeployment, Name: "payment-api", Namespace: "shop", Replicas: 1, ReadyReplicas: 0},
		Pods:     []PodFact{{Name: "payment-api-0", Namespace: "shop", Phase: "Pending"}},
		Containers: []ContainerFact{{
			Pod: "payment-api-0", Namespace: "shop", Name: "app", WaitingReason: "ImagePullBackOff",
		}},
		Events: []EventFact{{
			Name: "payment-api-0.1", Namespace: "shop", Reason: "Failed", InvolvedKind: "Pod", InvolvedName: "payment-api-0",
		}},
	}
	g := BuildGraph(snap)
	if len(g.Nodes) < 3 {
		t.Fatalf("expected workload, pod, container nodes, got %d", len(g.Nodes))
	}
	foundOwns := false
	for _, e := range g.Edges {
		if e.Relation == RelOwns {
			foundOwns = true
		}
	}
	if !foundOwns {
		t.Fatal("expected owns edge from deployment to pod")
	}
	var podStatus, ctrStatus string
	for _, n := range g.Nodes {
		if n.Kind == KindPod {
			podStatus = n.Status
		}
		if n.Kind == KindContainer {
			ctrStatus = n.Status
		}
	}
	if !strings.Contains(podStatus, "phase=Pending") || !strings.Contains(podStatus, "ready=") {
		t.Fatalf("pod status should distinguish phase vs readiness, got %q", podStatus)
	}
	if !strings.Contains(ctrStatus, "Waiting") || !strings.Contains(ctrStatus, "ImagePullBackOff") {
		t.Fatalf("container status should be Waiting(reason), got %q", ctrStatus)
	}
}
