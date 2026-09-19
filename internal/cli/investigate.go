package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aravinddudam/kubecone/internal/engine"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
	"github.com/aravinddudam/kubecone/internal/reporter"
)

func investigateCmd(f *flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "investigate RESOURCE",
		Short: "Collect evidence and diagnose a Kubernetes resource",
		Long: `Investigate a workload by collecting Kubernetes evidence, building a
graph, and running deterministic analyzers.

Examples:
  kubecone investigate deployment/payment-api
  kubecone investigate deploy/payment-api -n payments -o json
  kubecone investigate pod/payment-api-abc`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			timeout, err := time.ParseDuration(f.timeout)
			if err != nil {
				return fmt.Errorf("invalid --timeout: %w", err)
			}
			target, err := kubernetes.ParseTarget(args[0], f.namespace)
			if err != nil {
				return err
			}

			eng, err := engine.New(engine.Options{
				Kubeconfig: f.kubeconfig,
				Context:    f.context,
				Namespace:  f.namespace,
				LogTail:    f.logTail,
				Timeout:    timeout,
				AI:         f.ai,
				AIProvider: f.aiProvider,
			})
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			report, err := eng.Investigate(ctx, target)
			if err != nil {
				return err
			}
			return reporter.Write(cmd.OutOrStdout(), report, f.output)
		},
	}

	cmd.Flags().Int64Var(&f.logTail, "log-tail", 80, "log lines to collect per container")
	cmd.Flags().StringVar(&f.timeout, "timeout", "30s", "investigation timeout")
	cmd.Flags().BoolVar(&f.ai, "ai", false, "send ranked evidence to a configured AI provider (off by default)")
	cmd.Flags().StringVar(&f.aiProvider, "ai-provider", "", "ai provider: openai|anthropic|ollama")
	return cmd
}
