package cli

import (
	"errors"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/aravinddudam/kubecone/internal/version"
)

// ErrFinding is returned when --fail is set and a failure was diagnosed.
var ErrFinding = errors.New("failure diagnosed")

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, ErrFinding) {
		return 2
	}
	return 1
}

type flags struct {
	kubeconfig    string
	context       string
	namespace     string
	output        string
	logTail       int64
	timeout       string
	ai            bool
	aiProvider    string
	quiet         bool
	noColor       bool
	noGraph       bool
	fail          bool
	events        int
	allNamespaces bool
	unhealthy     bool
	kind          string
}

func New() *cobra.Command {
	f := &flags{
		output:  "text",
		logTail: 80,
		timeout: "30s",
		events:  8,
		kind:    "deployment",
	}

	root := &cobra.Command{
		Use:   "kubecone",
		Short: "Deterministic Kubernetes investigation CLI",
		Long: `KubeCone collects cluster evidence, builds an evidence graph, and
diagnoses common workload failures before any AI model is involved.

Commands:
  investigate      Diagnose one Deployment, Pod, or Service
  scan             List workloads in a namespace and flag failures
  pods             List pods and flag failures
  events           List recent Warning events
  nodes            List nodes
  namespaces       List namespaces
  current-context  Show kubeconfig context and API server
  explain          Describe a diagnosis code
  version          Print build information`,
		Example: `  kubecone pods -n banking-dev --unhealthy
  kubecone investigate deployment/customer-service -n banking-dev
  kubecone scan -n banking-dev --unhealthy
  kubecone events -n banking-dev
  kubecone explain IMAGE_PULL`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.String(),
	}

	root.SetVersionTemplate("kubecone {{.Version}}\n")
	root.PersistentFlags().StringVar(&f.kubeconfig, "kubeconfig", "", "path to kubeconfig (defaults to KUBECONFIG or ~/.kube/config)")
	root.PersistentFlags().StringVar(&f.context, "context", "", "kubeconfig context")
	root.PersistentFlags().StringVarP(&f.namespace, "namespace", "n", "", "namespace (defaults to the current context)")
	root.PersistentFlags().StringVarP(&f.output, "output", "o", "text", "output format: text|json")
	root.PersistentFlags().BoolVarP(&f.quiet, "quiet", "q", false, "print only the primary diagnosis")
	root.PersistentFlags().BoolVar(&f.noColor, "no-color", false, "disable ANSI color")
	root.PersistentFlags().BoolVar(&f.fail, "fail", false, "exit 2 when a failure is diagnosed (for CI)")

	root.AddCommand(investigateCmd(f))
	root.AddCommand(scanCmd(f))
	root.AddCommand(podsCmd(f))
	root.AddCommand(eventsCmd(f))
	root.AddCommand(nodesCmd(f))
	root.AddCommand(namespacesCmd(f))
	root.AddCommand(contextCmd(f))
	root.AddCommand(explainCmd())
	root.AddCommand(versionCmd())
	return root
}

func versionCmd() *cobra.Command {
	short := false
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print KubeCone version",
		Example: `  kubecone version
  kubecone version --short`,
		Run: func(cmd *cobra.Command, _ []string) {
			if short {
				cmd.Printf("%s\n", version.Version)
				return
			}
			cmd.Printf("kubecone %s\n", version.String())
			cmd.Printf("go:      %s\n", runtime.Version())
			cmd.Printf("os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
	cmd.Flags().BoolVar(&short, "short", false, "print the version number only")
	return cmd
}
