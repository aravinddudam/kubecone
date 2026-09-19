package evidence

import "strings"

// Correlate links events, containers, and owners so findings can cite evidence IDs.
func Correlate(snap Snapshot, finding Finding) Finding {
	ids := map[string]struct{}{}
	for _, id := range finding.EvidenceIDs {
		ids[id] = struct{}{}
	}

	res := finding.Resource
	for _, pod := range snap.Pods {
		podID := nodeID(KindPod, pod.Namespace, pod.Name)
		if strings.Contains(res, pod.Name) {
			ids[podID] = struct{}{}
		}
		for _, ev := range snap.Events {
			if ev.InvolvedName == pod.Name || strings.Contains(res, ev.InvolvedName) {
				if relatedEvent(ev, finding.Code) {
					ids[nodeID(KindEvent, ev.Namespace, ev.Name)] = struct{}{}
				}
			}
		}
	}
	for _, c := range snap.Containers {
		if strings.Contains(res, c.Pod) || strings.Contains(res, c.Name) {
			ids[nodeID(KindContainer, c.Namespace, c.Pod+"/"+c.Name)] = struct{}{}
		}
	}
	if snap.Workload != nil {
		ids[nodeID(snap.Workload.Kind, snap.Workload.Namespace, snap.Workload.Name)] = struct{}{}
	}

	finding.EvidenceIDs = make([]string, 0, len(ids))
	for id := range ids {
		finding.EvidenceIDs = append(finding.EvidenceIDs, id)
	}
	return finding
}

func relatedEvent(ev EventFact, code string) bool {
	r := strings.ToLower(ev.Reason)
	switch code {
	case CodeOOMKilled:
		return strings.Contains(r, "oom") || strings.Contains(strings.ToLower(ev.Message), "oom")
	case CodeCrashLoop:
		return strings.Contains(r, "backoff") || strings.Contains(r, "unhealthy") || r == "failed"
	case CodeImagePull:
		return strings.Contains(r, "image") || strings.Contains(r, "failed")
	case CodeFailedScheduling:
		return strings.Contains(r, "schedul")
	case CodeProbeFailed:
		return strings.Contains(r, "unhealthy") || strings.Contains(r, "probe")
	default:
		return ev.Type == "Warning"
	}
}
