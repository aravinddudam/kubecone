package engine

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/aravinddudam/kubecone/internal/kubernetes"
	"github.com/aravinddudam/kubecone/internal/reporter"
)

func TestListInventory(t *testing.T) {
	now := metav1.NewTime(time.Now())
	cs := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "banking-dev"}, Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive}},
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "ip-172-31-1-86.ec2.internal",
				Labels: map[string]string{"node-role.kubernetes.io/control-plane": "true"},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
				Addresses:  []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "172.31.1.86"}},
				NodeInfo:   corev1.NodeSystemInfo{KubeletVersion: "v1.36.4+k3s1"},
			},
		},
		&corev1.Event{
			ObjectMeta: metav1.ObjectMeta{Name: "pull", Namespace: "banking-dev"},
			Type:       corev1.EventTypeWarning,
			Reason:     "Failed",
			Message:    "Failed to pull image",
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "customer-service-1",
			},
			LastTimestamp: now,
			Count:         12,
		},
		&corev1.Event{
			ObjectMeta:     metav1.ObjectMeta{Name: "ok", Namespace: "banking-dev"},
			Type:           corev1.EventTypeNormal,
			Reason:         "Pulled",
			Message:        "Successfully pulled image",
			InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "ok"},
			LastTimestamp:  now,
		},
	)
	client := kubernetes.NewForInterface(cs, "banking-dev")

	nodes, err := ListNodes(context.Background(), client)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].Status != "Ready" || nodes[0].Roles != "control-plane" {
		t.Fatalf("nodes=%+v", nodes)
	}

	nss, err := ListNamespaces(context.Background(), client)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, ns := range nss {
		if ns.Name == "banking-dev" {
			found = true
		}
	}
	if !found {
		t.Fatalf("namespaces=%+v", nss)
	}

	events, err := ListEvents(context.Background(), client, EventOptions{Namespace: "banking-dev", Type: "warning", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Reason != "Failed" {
		t.Fatalf("events=%+v", events)
	}

	info := ContextInfo(client)
	if info.Context != "fake" || info.Namespace != "banking-dev" {
		t.Fatalf("context=%+v", info)
	}

	var buf bytes.Buffer
	if err := reporter.WriteNodes(&buf, nodes, "text"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "control-plane") {
		t.Fatalf("nodes table: %s", buf.String())
	}
}
