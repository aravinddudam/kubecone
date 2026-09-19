package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aravinddudam/kubecone/internal/engine"
	"github.com/aravinddudam/kubecone/internal/evidence"
	"github.com/aravinddudam/kubecone/internal/kubernetes"
	"github.com/aravinddudam/kubecone/internal/reporter"
)

func investigateCmd(f *flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "investigate RESOURCE",
		Aliases: []string{"inv", "diag"},
		Short:   "Collect evidence and diagnose a Kubernetes resource",
		Long: `Collect evidence and diagnose one workload.

RESOURCE is kind/name, or namespace/kind/name. A bare name is treated as a
Deployment. Supported kinds: pod, deploy, rs, svc, sts, ds, job, cronjob.

scan finds problems. investigate explains one object. explain documents a code.`,
		Example: `  kubecone investigate deployment/customer-service -n banking-dev
  kubecone investigate deploy/fraud-api -n banking-dev -o json
  kubecone investigate pod/customer-service-5859cfc476-bbkc6 -n banking-dev
  kubecone investigate banking-dev/deploy/customer-service --quiet
  kubecone investigate deploy/payment-api -n shop --fail --no-graph
  kubecone investigate deployment/customer-service -n banking-dev --ai`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			timeout, err := parseTimeout(f.timeout)
			if err != nil {
				return err
			}
			if f.ai && f.timeout == "30s" {
				timeout = 90 * time.Second
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
			return writeReport(cmd, f, report)
		},
	}

	cmd.Flags().Int64Var(&f.logTail, "log-tail", 80, "log lines to collect per container")
	cmd.Flags().StringVar(&f.timeout, "timeout", "30s", "investigation timeout")
	cmd.Flags().IntVar(&f.events, "events", 8, "recent events to print (0 hides the section)")
	cmd.Flags().BoolVar(&f.noGraph, "no-graph", false, "omit the evidence graph from text output")
	cmd.Flags().BoolVar(&f.ai, "ai", false, "after the rules diagnosis, ask OpenAI for cause and next steps")
	cmd.Flags().StringVar(&f.aiProvider, "ai-provider", "openai", "ai provider: openai (anthropic and ollama are still stubs)")
	return cmd
}

func writeReport(cmd *cobra.Command, f *flags, report *evidence.Report) error {
	if err := reporter.Write(cmd.OutOrStdout(), report, reporter.Options{
		Format:     f.output,
		Quiet:      f.quiet,
		NoGraph:    f.noGraph,
		NoColor:    f.noColor,
		HideAISkip: !f.ai,
		MaxEvents:  f.events,
	}); err != nil {
		return err
	}
	if f.fail && isFailure(report) {
		return ErrFinding
	}
	return nil
}

func isFailure(report *evidence.Report) bool {
	return report.Primary != nil && report.Primary.Code != "" && report.Primary.Code != evidence.CodeHealthy
}

func parseTimeout(raw string) (time.Duration, error) {
	timeout, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid --timeout %q: %w", raw, err)
	}
	if timeout <= 0 {
		return 0, fmt.Errorf("--timeout must be greater than 0")
	}
	return timeout, nil
}
