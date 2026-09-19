package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aravinddudam/kubecone/internal/analyzer"
)

func explainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain [CODE]",
		Short: "Explain a diagnosis code",
		Long: `Print what a KubeCone diagnosis code means, common causes, and what to do next.

scan finds problems. investigate explains one object. explain documents a code.

With no argument, lists every code v0.2 can emit.`,
		Example: `  kubecone explain
  kubecone explain IMAGE_PULL
  kubecone explain OOM_KILLED`,
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			out := make([]string, 0, len(analyzer.Codes()))
			for _, c := range analyzer.Codes() {
				out = append(out, c.Code)
			}
			return out, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if len(args) == 0 {
				fmt.Fprintln(out, "CODE                 MEANING")
				for _, c := range analyzer.Codes() {
					fmt.Fprintf(out, "%-20s %s\n", c.Code, c.Title)
				}
				fmt.Fprintln(out)
				fmt.Fprintln(out, `Use "kubecone explain CODE" for causes and the recommended next step.`)
				return nil
			}
			info, ok := analyzer.LookupCode(args[0])
			if !ok {
				return fmt.Errorf("unknown code %q (try kubecone explain)", args[0])
			}
			fmt.Fprintf(out, "%s\n\n", info.Code)
			fmt.Fprintf(out, "%s\n\n", info.Title)
			fmt.Fprintf(out, "%s\n", info.Summary)
			if len(info.Causes) > 0 {
				fmt.Fprintln(out)
				fmt.Fprintln(out, "Common causes:")
				for _, cause := range info.Causes {
					fmt.Fprintf(out, "  - %s\n", cause)
				}
			}
			if info.Next != "" {
				fmt.Fprintln(out)
				fmt.Fprintf(out, "Next:\n  %s\n", info.Next)
			}
			if info.Investigate != "" {
				fmt.Fprintln(out)
				fmt.Fprintln(out, "Investigate with:")
				fmt.Fprintf(out, "  %s\n", info.Investigate)
				fmt.Fprintln(out, "  kubecone scan -n <namespace> --unhealthy")
			}
			return nil
		},
	}
	return cmd
}
