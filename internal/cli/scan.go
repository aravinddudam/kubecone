package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aravinddudam/kubecone/internal/engine"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
	"github.com/aravinddudam/kubecone/internal/reporter"
)

func scanCmd(f *flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scan",
		Aliases: []string{"ls"},
		Short:   "Scan a namespace for unhealthy workloads",
		Long: `List Deployments (and optionally StatefulSets / DaemonSets) and run the
same deterministic rules used by investigate. Use this to find what to
investigate next.

By default scan looks at Deployments in the current namespace.`,
		Example: `  kubecone scan -n banking-dev
  kubecone scan -n banking-dev --unhealthy
  kubecone scan -A --kind all
  kubecone scan -n kube-system --kind daemonset -o json
  kubecone scan -n banking-dev --fail --unhealthy`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			timeout, err := parseTimeout(f.timeout)
			if err != nil {
				return err
			}
			client, err := kubernetes.New(kubernetes.Options{
				Kubeconfig: f.kubeconfig,
				Context:    f.context,
				Namespace:  f.namespace,
			})
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
			if strings.EqualFold(f.output, "json") || !f.quiet {
				if err := reporter.WriteScan(cmd.OutOrStdout(), items, f.output); err != nil {
					return err
				}
			} else {
				for _, item := range items {
					if !item.Unhealthy() {
						continue
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s/%s\t%s\n", item.Code, item.Kind, item.Name, item.Namespace)
				}
			}
			if f.fail {
				for _, item := range items {
					if item.Unhealthy() {
						return ErrFinding
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&f.allNamespaces, "all-namespaces", "A", false, "scan every namespace")
	cmd.Flags().BoolVar(&f.unhealthy, "unhealthy", false, "show only workloads with a failure code")
	cmd.Flags().StringVar(&f.kind, "kind", "deployment", "workload kinds: deployment|statefulset|daemonset|all")
	cmd.Flags().StringVar(&f.timeout, "timeout", "30s", "scan timeout")
	return cmd
}
