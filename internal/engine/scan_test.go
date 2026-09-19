package engine

import (
	"context"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/aravinddudam/kubecone/internal/evidence"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
)

func TestScanImagePull(t *testing.T) {
	replicas := int32(1)
	cs := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "customer-service", Namespace: "banking-dev"},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "customer-service"}},
			},
			Status: appsv1.DeploymentStatus{ReadyReplicas: 0},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "customer-service-1",
				Namespace: "banking-dev",
				Labels:    map[string]string{"app": "customer-service"},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
				ContainerStatuses: []corev1.ContainerStatus{{
					Name:  "app",
					Image: "707405801682.dkr.ecr.us-east-1.amazonaws.com/app-dev/customer-service:d54a1c7",
					State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{
						Reason:  "ImagePullBackOff",
						Message: `failed to resolve reference`,
					}},
				}},
			},
		},
	)
	client := kubernetes.NewForInterface(cs, "banking-dev")
	items, err := Scan(context.Background(), client, ScanOptions{Namespace: "banking-dev"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items=%d", len(items))
	}
	if items[0].Code != evidence.CodeImagePull {
		t.Fatalf("code=%s status=%s", items[0].Code, items[0].Status)
	}
	if items[0].Status != "ImagePullBackOff" {
		t.Fatalf("status=%s", items[0].Status)
	}
}

func TestParseScanKinds(t *testing.T) {
	got := strings.Join(ParseScanKinds("all"), ",")
	if got != "Deployment,StatefulSet,DaemonSet" {
		t.Fatalf("all=%s", got)
	}
	if ParseScanKinds("sts")[0] != "StatefulSet" {
		t.Fatalf("sts=%v", ParseScanKinds("sts"))
	}
}
