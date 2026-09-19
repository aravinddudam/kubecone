package engine

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aravinddudam/kubecone/internal/evidence"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
)

type EventOptions struct {
	Namespace     string
	AllNamespaces bool
	Type          string
	Limit         int
}

func ListEvents(ctx context.Context, client *kubernetes.Client, opts EventOptions) ([]evidence.EventItem, error) {
	ns := namespaceOrAll(client, opts.Namespace, opts.AllNamespaces)

	list, err := client.ListEvents(ctx, ns)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	want := strings.ToLower(strings.TrimSpace(opts.Type))
	if want == "" {
		want = "warning"
	}

	type dated struct {
		item evidence.EventItem
		at   time.Time
	}
	datedItems := make([]dated, 0, len(list.Items))
	for i := range list.Items {
		ev := list.Items[i]
		if want != "all" && !strings.EqualFold(ev.Type, want) {
			continue
		}
		at := eventTime(ev)
		datedItems = append(datedItems, dated{
			at: at,
			item: evidence.EventItem{
				Namespace: ev.Namespace,
				Type:      ev.Type,
				Reason:    ev.Reason,
				Object:    ev.InvolvedObject.Kind + "/" + ev.InvolvedObject.Name,
				Message:   strings.TrimSpace(ev.Message),
				Count:     ev.Count,
				Age:       ageSince(at),
			},
		})
	}

	sort.Slice(datedItems, func(i, j int) bool {
		return datedItems[i].at.After(datedItems[j].at)
	})
	limit := opts.Limit
	if limit <= 0 || limit > len(datedItems) {
		limit = len(datedItems)
	}
	items := make([]evidence.EventItem, 0, limit)
	for i := 0; i < limit; i++ {
		items = append(items, datedItems[i].item)
	}
	return items, nil
}

func ListNodes(ctx context.Context, client *kubernetes.Client) ([]evidence.NodeItem, error) {
	list, err := client.ListNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	items := make([]evidence.NodeItem, 0, len(list.Items))
	for i := range list.Items {
		n := list.Items[i]
		items = append(items, evidence.NodeItem{
			Name:       n.Name,
			Status:     nodeReady(n),
			Roles:      nodeRoles(n),
			Version:    n.Status.NodeInfo.KubeletVersion,
			InternalIP: nodeIP(n, corev1.NodeInternalIP),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}

func ListNamespaces(ctx context.Context, client *kubernetes.Client) ([]evidence.NamespaceItem, error) {
	list, err := client.ListNamespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}
	items := make([]evidence.NamespaceItem, 0, len(list.Items))
	for i := range list.Items {
		ns := list.Items[i]
		items = append(items, evidence.NamespaceItem{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}

func ContextInfo(client *kubernetes.Client) evidence.ContextInfo {
	info := evidence.ContextInfo{
		Context:   client.Context,
		Namespace: client.Namespace,
	}
	if client.REST != nil {
		info.Server = client.REST.Host
	}
	return info
}

func namespaceOrAll(client *kubernetes.Client, namespace string, all bool) string {
	if all {
		return metav1.NamespaceAll
	}
	if namespace != "" {
		return namespace
	}
	return client.Namespace
}

func eventTime(ev corev1.Event) time.Time {
	if !ev.LastTimestamp.Time.IsZero() {
		return ev.LastTimestamp.Time
	}
	if !ev.EventTime.Time.IsZero() {
		return ev.EventTime.Time
	}
	return ev.CreationTimestamp.Time
}

func ageSince(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func nodeReady(n corev1.Node) string {
	for _, c := range n.Status.Conditions {
		if c.Type == corev1.NodeReady {
			if c.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

func nodeRoles(n corev1.Node) string {
	seen := map[string]struct{}{}
	var roles []string
	add := func(role string) {
		if role == "" {
			return
		}
		if _, ok := seen[role]; ok {
			return
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	for k, v := range n.Labels {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") {
			add(strings.TrimPrefix(k, "node-role.kubernetes.io/"))
		}
		if k == "kubernetes.io/role" {
			add(v)
		}
	}
	if len(roles) == 0 {
		return "<none>"
	}
	sort.Strings(roles)
	return strings.Join(roles, ",")
}

func nodeIP(n corev1.Node, kind corev1.NodeAddressType) string {
	for _, a := range n.Status.Addresses {
		if a.Type == kind {
			return a.Address
		}
	}
	return ""
}
