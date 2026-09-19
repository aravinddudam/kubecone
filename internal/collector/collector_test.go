package collector

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/aravinddudam/kubecone/internal/kubernetes"
)

func TestCollectDeployment(t *testing.T) {
	replicas := int32(1)
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "payment-api", Namespace: "shop"},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "payment-api"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "payment-api"}},
			},
		},
		Status: appsv1.DeploymentStatus{Replicas: 1, UnavailableReplicas: 1},
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "payment-api-0",
			Namespace: "shop",
			Labels:    map[string]string{"app": "payment-api"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "app",
				Image: "app:1",
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("32Mi")},
				},
			}},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:         "app",
				RestartCount: 3,
				State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
				},
				LastTerminationState: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{Reason: "OOMKilled", ExitCode: 137},
				},
			}},
		},
	}
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{Name: "payment-api-0.1", Namespace: "shop"},
		InvolvedObject: corev1.ObjectReference{
			Kind: "Pod", Name: "payment-api-0", Namespace: "shop",
		},
		Reason:  "BackOff",
		Message: "Back-off restarting failed container",
		Type:    "Warning",
		Count:   5,
	}

	client := kubernetes.NewForInterface(fake.NewSimpleClientset(deploy, pod, event), "shop")
	col := New(client, Options{LogTail: 20})
	snap, err := col.Collect(context.Background(), kubernetes.Target{
		Kind: "Deployment", Name: "payment-api", Namespace: "shop",
	})
	if err != nil {
		t.Fatal(err)
	}
	if snap.Workload == nil || snap.Workload.Name != "payment-api" {
		t.Fatalf("missing workload: %+v", snap.Workload)
	}
	if len(snap.Pods) != 1 {
		t.Fatalf("pods=%d", len(snap.Pods))
	}
	if len(snap.Containers) != 1 || snap.Containers[0].LastTerminatedReason != "OOMKilled" {
		t.Fatalf("containers=%+v", snap.Containers)
	}
	if len(snap.Events) != 1 {
		t.Fatalf("events=%+v", snap.Events)
	}
}
