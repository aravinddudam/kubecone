package collector

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func podFacts(pods []corev1.Pod) []evidence.PodFact {
	out := make([]evidence.PodFact, 0, len(pods))
	for _, p := range pods {
		fact := evidence.PodFact{
			Name:      p.Name,
			Namespace: p.Namespace,
			Phase:     string(p.Status.Phase),
			Reason:    p.Status.Reason,
			Message:   p.Status.Message,
			Node:      p.Spec.NodeName,
			QOS:       string(p.Status.QOSClass),
		}
		if kind, name := controllerOwner(p); name != "" {
			fact.OwnerKind = kind
			fact.OwnerName = name
		}
		for _, cond := range p.Status.Conditions {
			fact.Conditions = append(fact.Conditions, evidence.Condition{
				Type:    string(cond.Type),
				Status:  string(cond.Status),
				Reason:  cond.Reason,
				Message: cond.Message,
			})
		}
		out = append(out, fact)
	}
	return out
}

func controllerOwner(p corev1.Pod) (string, string) {
	for _, ref := range p.OwnerReferences {
		if ref.Controller != nil && *ref.Controller {
			return ref.Kind, ref.Name
		}
	}
	return "", ""
}
