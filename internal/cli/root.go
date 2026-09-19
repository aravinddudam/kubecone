package cli

import (
	"github.com/spf13/cobra"

	"github.com/aravinddudam/kubecone/internal/version"
)

type flags struct {
	kubeconfig string
	context    string
	namespace  string
	output     string
	logTail    int64
	timeout    string
	ai         bool
	aiProvider string
}

func New() *cobra.Command {
	f := &flags{
		output:  "text",
		logTail: 80,
		timeout: "30s",
	}

	root := &cobra.Command{
		Use:   "kubecone",
		Short: "Deterministic Kubernetes investigation CLI",
		Long: `KubeCone collects cluster evidence, builds an evidence graph, and
diagnoses common workload failures before any AI model is involved.

  kubecone investigate deployment/payment-api`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&f.kubeconfig, "kubeconfig", "", "path to kubeconfig")
	root.PersistentFlags().StringVar(&f.context, "context", "", "kubeconfig context")
	root.PersistentFlags().StringVarP(&f.namespace, "namespace", "n", "", "namespace (defaults to the current context)")
	root.PersistentFlags().StringVarP(&f.output, "output", "o", "text", "output format: text|json")

	root.AddCommand(investigateCmd(f))
	root.AddCommand(versionCmd())
	return root
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print KubeCone version",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Printf("kubecone %s (%s)\n", version.Version, version.Commit)
		},
	}
}
