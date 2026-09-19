package evidence

import "fmt"

func nodeID(kind, namespace, name string) string {
	if namespace == "" {
		return kind + "/" + name
	}
	return kind + "/" + namespace + "/" + name
}

func addNode(g *Graph, seen map[string]struct{}, n Node) {
	if _, ok := seen[n.ID]; ok {
		return
	}
	seen[n.ID] = struct{}{}
	g.Nodes = append(g.Nodes, n)
}

func addEdge(g *Graph, from, to, rel string) {
	if from == "" || to == "" {
		return
	}
	g.Edges = append(g.Edges, Edge{From: from, To: to, Relation: rel})
}

// BuildGraph turns a collected snapshot into an evidence graph.
func BuildGraph(snap Snapshot) Graph {
	g := Graph{}
	seen := map[string]struct{}{}

	if snap.Workload != nil {
		w := snap.Workload
		id := nodeID(w.Kind, w.Namespace, w.Name)
		addNode(&g, seen, Node{
			ID:        id,
			Kind:      w.Kind,
			Name:      w.Name,
			Namespace: w.Namespace,
			Status:    fmt.Sprintf("ready=%d/%d available=%d unavailable=%d", w.ReadyReplicas, w.Replicas, w.Available, w.Unavailable),
		})
	}

	for _, svc := range snap.Services {
		id := nodeID(KindService, svc.Namespace, svc.Name)
		addNode(&g, seen, Node{
			ID:        id,
			Kind:      KindService,
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Status:    svc.Type,
		})
	}

	for _, pod := range snap.Pods {
		id := nodeID(KindPod, pod.Namespace, pod.Name)
		addNode(&g, seen, Node{
			ID:        id,
			Kind:      KindPod,
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    podGraphStatus(pod, snap.Containers),
			Attributes: map[string]string{
				"node":   pod.Node,
				"reason": pod.Reason,
				"phase":  pod.Phase,
			},
		})
		if snap.Workload != nil {
			addEdge(&g, nodeID(snap.Workload.Kind, snap.Workload.Namespace, snap.Workload.Name), id, RelOwns)
		} else if pod.OwnerName != "" {
			addEdge(&g, nodeID(pod.OwnerKind, pod.Namespace, pod.OwnerName), id, RelOwns)
		}
		if pod.Node != "" {
			nid := nodeID(KindNode, "", pod.Node)
			addNode(&g, seen, Node{ID: nid, Kind: KindNode, Name: pod.Node})
			addEdge(&g, id, nid, RelRunsOn)
		}
	}

	for _, n := range snap.Nodes {
		addNode(&g, seen, Node{
			ID:     nodeID(KindNode, "", n.Name),
			Kind:   KindNode,
			Name:   n.Name,
			Status: readyStatus(n.Ready),
		})
	}

	for _, c := range snap.Containers {
		id := nodeID(KindContainer, c.Namespace, c.Pod+"/"+c.Name)
		addNode(&g, seen, Node{
			ID:        id,
			Kind:      KindContainer,
			Name:      c.Name,
			Namespace: c.Namespace,
			Status:    containerGraphStatus(c),
			Attributes: map[string]string{
				"image":         c.Image,
				"restarts":      fmt.Sprintf("%d", c.RestartCount),
				"memoryLimit":   c.MemoryLimit,
				"memoryRequest": c.MemoryRequest,
			},
		})
		addEdge(&g, nodeID(KindPod, c.Namespace, c.Pod), id, RelContains)
	}

	for _, ev := range snap.Events {
		id := nodeID(KindEvent, ev.Namespace, ev.Name)
		addNode(&g, seen, Node{
			ID:        id,
			Kind:      KindEvent,
			Name:      ev.Reason,
			Namespace: ev.Namespace,
			Status:    ev.Type,
			Attributes: map[string]string{
				"message": ev.Message,
				"count":   fmt.Sprintf("%d", ev.Count),
			},
		})
		if ev.InvolvedName != "" {
			addEdge(&g, id, nodeID(ev.InvolvedKind, ev.Namespace, ev.InvolvedName), RelReports)
		}
	}

	return g
}

func readyStatus(ready bool) string {
	if ready {
		return "Ready"
	}
	return "NotReady"
}

func podGraphStatus(pod PodFact, containers []ContainerFact) string {
	var ready, total int
	waiting := ""
	for _, c := range containers {
		if c.Pod != pod.Name {
			continue
		}
		total++
		if c.Ready {
			ready++
		}
		if waiting == "" && c.WaitingReason != "" {
			waiting = c.WaitingReason
		}
	}
	s := fmt.Sprintf("phase=%s ready=%d/%d", valueOr(pod.Phase, "Unknown"), ready, total)
	if waiting != "" {
		s += " waiting=" + waiting
	}
	return s
}

func containerGraphStatus(c ContainerFact) string {
	if c.WaitingReason != "" {
		return "Waiting (" + c.WaitingReason + ")"
	}
	if c.TerminatedReason != "" {
		return "Terminated (" + c.TerminatedReason + ")"
	}
	if c.Ready {
		return "Ready"
	}
	if c.Started {
		return "Running (not ready)"
	}
	return "running"
}

func valueOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
