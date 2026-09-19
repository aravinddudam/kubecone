package kubernetes

import (
	"fmt"
	"strings"
)

// Target is a Kubernetes object KubeCone should investigate.
type Target struct {
	Kind      string
	Name      string
	Namespace string
}

func (t Target) String() string {
	if t.Namespace == "" {
		return t.Kind + "/" + t.Name
	}
	return t.Namespace + "/" + t.Kind + "/" + t.Name
}

var kindAliases = map[string]string{
	"po":           "Pod",
	"pod":          "Pod",
	"pods":         "Pod",
	"deploy":       "Deployment",
	"deployment":   "Deployment",
	"deployments":  "Deployment",
	"rs":           "ReplicaSet",
	"replicaset":   "ReplicaSet",
	"replicasets":  "ReplicaSet",
	"svc":          "Service",
	"service":      "Service",
	"services":     "Service",
	"sts":          "StatefulSet",
	"statefulset":  "StatefulSet",
	"statefulsets": "StatefulSet",
	"ds":           "DaemonSet",
	"daemonset":    "DaemonSet",
	"job":          "Job",
	"jobs":         "Job",
	"cj":           "CronJob",
	"cronjob":      "CronJob",
}

// ParseTarget turns "deployment/payment-api" (or "deploy/payment-api") into a Target.
func ParseTarget(raw, defaultNamespace string) (Target, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Target{}, fmt.Errorf("resource is required, e.g. deployment/payment-api")
	}

	ns := defaultNamespace
	kind := ""
	name := ""

	parts := strings.Split(raw, "/")
	switch len(parts) {
	case 1:
		kind = "Deployment"
		name = parts[0]
	case 2:
		kind = parts[0]
		name = parts[1]
	case 3:
		ns = parts[0]
		kind = parts[1]
		name = parts[2]
	default:
		return Target{}, fmt.Errorf("invalid resource %q (use kind/name)", raw)
	}

	canonical, ok := kindAliases[strings.ToLower(kind)]
	if !ok {
		return Target{}, fmt.Errorf("unsupported kind %q", kind)
	}
	if name == "" {
		return Target{}, fmt.Errorf("resource name is required")
	}
	if ns == "" {
		ns = "default"
	}
	return Target{Kind: canonical, Name: name, Namespace: ns}, nil
}
