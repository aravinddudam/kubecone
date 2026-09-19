package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/aravinddudam/kubecone/internal/engine"
	"github.com/aravinddudam/kubecone/internal/evidence"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
	"github.com/aravinddudam/kubecone/internal/reporter"
)

func podsCmd(f *flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pods",
		Aliases: []string{"po", "pod"},
		Short:   "List pods and attach a failure code",
		Long: `List pods in a namespace (or every namespace) and run the same
deterministic rules used by investigate. Use this when scan misses a
broken pod because it is not owned by a Deployment.`,
		Example: `  kubecone pods -n banking-dev --unhealthy
  kubecone pods -A --unhealthy
  kubecone pods -n argocd -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, timeout, err := f.connect()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			items, err := engine.ScanPods(ctx, client, engine.ScanOptions{
				Namespace:     f.namespace,
				AllNamespaces: f.allNamespaces,
				UnhealthyOnly: f.unhealthy,
			})
			if err != nil {
				return err
			}
			if err := reporter.WriteScanMeta(cmd.OutOrStdout(), items, f.output, reporter.ScanMeta{
				Title:         "KubeCone pods",
				Context:       client.Context,
				Namespace:     nsOrAll(f, client.Namespace),
				UnhealthyOnly: f.unhealthy,
				EmptyHint:     "No matching pods.",
			}); err != nil {
				return err
			}
			return failIfUnhealthy(f.fail, items)
		},
	}
	cmd.Flags().BoolVarP(&f.allNamespaces, "all-namespaces", "A", false, "list pods in every namespace")
	cmd.Flags().BoolVar(&f.unhealthy, "unhealthy", false, "show only pods with a failure or non-ready status")
	cmd.Flags().StringVar(&f.timeout, "timeout", "30s", "list timeout")
	return cmd
}

func eventsCmd(f *flags) *cobra.Command {
	limit := 20
	eventType := "warning"
	cmd := &cobra.Command{
		Use:     "events",
		Aliases: []string{"ev"},
		Short:   "List recent cluster events",
		Long: `Print Kubernetes events, newest first. Default is Warning events
in the current namespace.`,
		Example: `  kubecone events -n banking-dev
  kubecone events -A --type all
  kubecone events -n argocd --limit 50 -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, timeout, err := f.connect()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			items, err := engine.ListEvents(ctx, client, engine.EventOptions{
				Namespace:     f.namespace,
				AllNamespaces: f.allNamespaces,
				Type:          eventType,
				Limit:         limit,
			})
			if err != nil {
				return err
			}
			return reporter.WriteEvents(cmd.OutOrStdout(), items, f.output)
		},
	}
	cmd.Flags().BoolVarP(&f.allNamespaces, "all-namespaces", "A", false, "list events in every namespace")
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum events to print")
	cmd.Flags().StringVar(&eventType, "type", "warning", "event type: warning|normal|all")
	cmd.Flags().StringVar(&f.timeout, "timeout", "30s", "list timeout")
	return cmd
}

func nodesCmd(f *flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nodes",
		Aliases: []string{"node", "no"},
		Short:   "List nodes and Ready status",
		Example: `  kubecone nodes
  kubecone nodes -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, timeout, err := f.connect()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			items, err := engine.ListNodes(ctx, client)
			if err != nil {
				return err
			}
			return reporter.WriteNodes(cmd.OutOrStdout(), items, f.output)
		},
	}
	cmd.Flags().StringVar(&f.timeout, "timeout", "30s", "list timeout")
	return cmd
}

func namespacesCmd(f *flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "namespaces",
		Aliases: []string{"ns", "namespace"},
		Short:   "List namespaces",
		Example: `  kubecone namespaces
  kubecone ns -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, timeout, err := f.connect()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			items, err := engine.ListNamespaces(ctx, client)
			if err != nil {
				return err
			}
			return reporter.WriteNamespaces(cmd.OutOrStdout(), items, f.output)
		},
	}
	cmd.Flags().StringVar(&f.timeout, "timeout", "30s", "list timeout")
	return cmd
}

func contextCmd(f *flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "current-context",
		Aliases: []string{"ctx"},
		Short:   "Show the kubeconfig context KubeCone will use",
		Long: `Print the kubeconfig context, namespace, and API server.
Use this to confirm kubecone is talking to the same cluster as kubectl.`,
		Example: `  kubecone current-context
  kubecone ctx --kubeconfig /etc/rancher/k3s/k3s.yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := f.client()
			if err != nil {
				return err
			}
			return reporter.WriteContext(cmd.OutOrStdout(), engine.ContextInfo(client), f.output)
		},
	}
	return cmd
}

func (f *flags) client() (*kubernetes.Client, error) {
	return kubernetes.New(kubernetes.Options{
		Kubeconfig: f.kubeconfig,
		Context:    f.context,
		Namespace:  f.namespace,
	})
}

func (f *flags) connect() (*kubernetes.Client, time.Duration, error) {
	timeout, err := parseTimeout(f.timeout)
	if err != nil {
		return nil, 0, err
	}
	client, err := f.client()
	if err != nil {
		return nil, 0, err
	}
	return client, timeout, nil
}

func failIfUnhealthy(fail bool, items []evidence.ScanItem) error {
	if !fail {
		return nil
	}
	for _, item := range items {
		if item.Unhealthy() {
			return ErrFinding
		}
	}
	return nil
}

func nsOrAll(f *flags, fallback string) string {
	if f.allNamespaces {
		return "(all)"
	}
	if f.namespace != "" {
		return f.namespace
	}
	return fallback
}
