package analyzer

import "github.com/aravinddudam/kubecone/internal/evidence"

// CodeInfo is the human-readable catalog behind `kubecone explain`.
type CodeInfo struct {
	Code    string
	Title   string
	Summary string
	Next    string
}

func Codes() []CodeInfo {
	return []CodeInfo{
		{
			Code:    evidence.CodeImagePull,
			Title:   "Image cannot be pulled",
			Summary: "The kubelet could not fetch the container image (ErrImagePull / ImagePullBackOff).",
			Next:    "Check the image name and tag, registry reachability, and credentials. For Amazon ECR, confirm IAM/IRSA can pull from the repository.",
		},
		{
			Code:    evidence.CodeFailedScheduling,
			Title:   "Pod cannot be scheduled",
			Summary: "The pod never bound to a node (Pending / Unschedulable).",
			Next:    "Compare requests to node capacity, then check nodeSelector, affinity, taints, and PVC binding.",
		},
		{
			Code:    evidence.CodeOOMKilled,
			Title:   "Container killed because it ran out of memory",
			Summary: "The container exceeded its memory limit and was OOMKilled (often followed by CrashLoopBackOff).",
			Next:    "Raise the memory limit, or reduce the process working set.",
		},
		{
			Code:    evidence.CodeProbeFailed,
			Title:   "Readiness or liveness probe is failing",
			Summary: "The container is running but Kubernetes probes are failing, so the pod is not Ready or is being restarted.",
			Next:    "Confirm the probe path, port, and initialDelaySeconds match what the process actually serves.",
		},
		{
			Code:    evidence.CodeCrashLoop,
			Title:   "Container is crash looping",
			Summary: "The container starts, exits, and Kubernetes is backing off restarts.",
			Next:    "Inspect the last exit code and logs. Fix the command, config, or dependency the process needs to stay running.",
		},
		{
			Code:    evidence.CodeHealthy,
			Title:   "No deterministic failure found",
			Summary: "Collected evidence did not match OOM, crash loop, image pull, scheduling, or probe rules.",
			Next:    "If the workload is still unhealthy, inspect events and logs, or pass a more specific pod target.",
		},
	}
}

func LookupCode(code string) (CodeInfo, bool) {
	for _, c := range Codes() {
		if c.Code == code {
			return c, true
		}
	}
	return CodeInfo{}, false
}
