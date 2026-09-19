package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/aravinddudam/kubecone/internal/engine"
	"github.com/aravinddudam/kubecone/internal/reporter"
)

func scanCmd(f *flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scan",
		Aliases: []string{"ls"},
		Short:   "Scan a namespace for unhealthy workloads",
		Long: `Find problems in a namespace.

scan lists Deployments (and optionally other workloads) and attaches the same
failure codes investigate uses. Use --unhealthy to hide ready workloads.

Then run investigate on one resource, and explain on the diagnosis code.`,
		Example: `  kubecone scan -n banking-dev
  kubecone scan -n banking-dev --unhealthy
  kubecone scan -A --kind all
  kubecone scan -n kube-system --kind daemonset -o json
  kubecone scan -n banking-dev --fail --unhealthy`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, timeout, err := f.connect()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			items, err := engine.Scan(ctx, client, engine.ScanOptions{
				Namespace:     f.namespace,
				AllNamespaces: f.allNamespaces,
				UnhealthyOnly: f.unhealthy,
				Kinds:         engine.ParseScanKinds(f.kind),
			})
			if err != nil {
				return err
			}
			ns := f.namespace
			if f.allNamespaces {
				ns = "(all)"
			} else if ns == "" {
				ns = client.Namespace
			}
			if err := reporter.WriteScanMeta(cmd.OutOrStdout(), items, f.output, reporter.ScanMeta{
				Title:         "KubeCone scan",
				Context:       client.Context,
				Namespace:     ns,
				UnhealthyOnly: f.unhealthy,
				EmptyHint:     "No matching workloads.",
			}); err != nil {
				return err
			}
			return failIfUnhealthy(f.fail, items)
		},
	}

	cmd.Flags().BoolVarP(&f.allNamespaces, "all-namespaces", "A", false, "scan every namespace")
	cmd.Flags().BoolVar(&f.unhealthy, "unhealthy", false, "show only workloads with a failure code")
	cmd.Flags().StringVar(&f.kind, "kind", "deployment", "workload kinds: deployment|statefulset|daemonset|all")
	cmd.Flags().StringVar(&f.timeout, "timeout", "30s", "scan timeout")
	return cmd
}
