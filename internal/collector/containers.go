package collector

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func ContainerFacts(pods []corev1.Pod) []evidence.ContainerFact {
	return containerFacts(pods)
}

func containerFacts(pods []corev1.Pod) []evidence.ContainerFact {
	var out []evidence.ContainerFact
	for _, pod := range pods {
		limits := resourceLookup(pod)
		for _, cs := range append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...) {
			fact := evidence.ContainerFact{
				Pod:          pod.Name,
				Namespace:    pod.Namespace,
				Name:         cs.Name,
				Image:        cs.Image,
				Ready:        cs.Ready,
				RestartCount: cs.RestartCount,
				Started:      cs.Started != nil && *cs.Started,
			}
			if res, ok := limits[cs.Name]; ok {
				fact.MemoryRequest = res.memReq
				fact.MemoryLimit = res.memLim
				fact.CPURequest = res.cpuReq
				fact.CPULimit = res.cpuLim
			}
			if w := cs.State.Waiting; w != nil {
				fact.WaitingReason = w.Reason
				fact.WaitingMessage = w.Message
			}
			if t := cs.State.Terminated; t != nil {
				fact.TerminatedReason = t.Reason
				fact.TerminatedMessage = t.Message
				code := t.ExitCode
				fact.TerminatedExitCode = &code
			}
			if t := cs.LastTerminationState.Terminated; t != nil {
				fact.LastTerminatedReason = t.Reason
				fact.LastTerminatedMessage = t.Message
				code := t.ExitCode
				fact.LastTerminatedExitCode = &code
			}
			out = append(out, fact)
		}
	}
	return out
}

type resources struct {
	memReq, memLim, cpuReq, cpuLim string
}

func resourceLookup(pod corev1.Pod) map[string]resources {
	out := map[string]resources{}
	add := func(c corev1.Container) {
		r := resources{}
		if c.Resources.Requests != nil {
			if q := c.Resources.Requests[corev1.ResourceMemory]; !q.IsZero() {
				r.memReq = q.String()
			}
			if q := c.Resources.Requests[corev1.ResourceCPU]; !q.IsZero() {
				r.cpuReq = q.String()
			}
		}
		if c.Resources.Limits != nil {
			if q := c.Resources.Limits[corev1.ResourceMemory]; !q.IsZero() {
				r.memLim = q.String()
			}
			if q := c.Resources.Limits[corev1.ResourceCPU]; !q.IsZero() {
				r.cpuLim = q.String()
			}
		}
		out[c.Name] = r
	}
	for _, c := range pod.Spec.Containers {
		add(c)
	}
	for _, c := range pod.Spec.InitContainers {
		add(c)
	}
	return out
}
