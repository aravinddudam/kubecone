package analyzer

import (
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

// CodeInfo is the human-readable catalog behind `kubecone explain`.
type CodeInfo struct {
	Code        string
	Title       string
	Summary     string
	Causes      []string
	Next        string
	Investigate string
}

func Codes() []CodeInfo {
	return []CodeInfo{
		{
			Code:    evidence.CodeImagePull,
			Title:   "Image cannot be pulled",
			Summary: "The container runtime could not download the configured container image (ErrImagePull / ImagePullBackOff).",
			Causes: []string{
				"repository or image tag does not exist",
				"ECR/registry authentication or IAM failure",
				"insufficient registry permissions",
				"node cannot reach the registry (network, security group, or private endpoint)",
				"DNS cannot resolve the registry hostname",
				"image architecture/platform mismatch",
			},
			Next:        "Read the container-runtime error on the finding. Do not assume imagePullSecret unless the message says unauthorized/denied.",
			Investigate: "kubecone investigate <pod-or-workload>",
		},
		{
			Code:    evidence.CodeFailedScheduling,
			Title:   "Pod cannot be scheduled",
			Summary: "The pod never bound to a node (Pending / Unschedulable).",
			Causes: []string{
				"CPU/memory requests exceed node capacity",
				"nodeSelector, affinity, or taints do not match",
				"PVC cannot bind",
			},
			Next:        "Compare requests to node capacity, then check nodeSelector, affinity, taints, and PVC binding.",
			Investigate: "kubecone investigate <pod-or-workload>",
		},
		{
			Code:    evidence.CodeOOMKilled,
			Title:   "Container killed because it ran out of memory",
			Summary: "The container exceeded its memory limit and was OOMKilled (often followed by CrashLoopBackOff).",
			Causes: []string{
				"memory limit too low for the process working set",
				"memory leak",
			},
			Next:        "Raise the memory limit, or reduce the process working set.",
			Investigate: "kubecone investigate <pod-or-workload>",
		},
		{
			Code:    evidence.CodeProbeFailed,
			Title:   "Readiness or liveness probe is failing",
			Summary: "The container is running but Kubernetes probes are failing, so the pod is not Ready or is being restarted.",
			Causes: []string{
				"probe path or port does not match the process",
				"initialDelaySeconds too short",
				"app not listening yet",
			},
			Next:        "Confirm the probe path, port, and initialDelaySeconds match what the process actually serves.",
			Investigate: "kubecone investigate <pod-or-workload>",
		},
		{
			Code:    evidence.CodeCrashLoop,
			Title:   "Container is crash looping",
			Summary: "The container starts, exits, and Kubernetes is backing off restarts.",
			Causes: []string{
				"process exits non-zero",
				"missing config or dependency",
				"bad command or args",
			},
			Next:        "Inspect the last exit code and logs. Fix the command, config, or dependency the process needs to stay running.",
			Investigate: "kubecone investigate <pod-or-workload>",
		},
		{
			Code:    evidence.CodeHealthy,
			Title:   "No deterministic failure found",
			Summary: "Collected evidence did not match OOM, crash loop, image pull, scheduling, or probe rules.",
			Next:    "If the workload is still unhealthy, inspect events and logs, or pass a more specific pod target.",
		},
	}
}

func NormalizeCode(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

func LookupCode(code string) (CodeInfo, bool) {
	want := NormalizeCode(code)
	for _, c := range Codes() {
		if c.Code == want {
			return c, true
		}
	}
	return CodeInfo{}, false
}
