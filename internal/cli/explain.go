package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aravinddudam/kubecone/internal/analyzer"
)

func explainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain [CODE]",
		Short: "Explain a diagnosis code",
		Long: `Print what a KubeCone diagnosis code means and what to do next.

With no argument, lists every code v0.1 can emit.`,
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
			if len(args) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "CODE                 MEANING")
				for _, c := range analyzer.Codes() {
					fmt.Fprintf(cmd.OutOrStdout(), "%-20s %s\n", c.Code, c.Title)
				}
				fmt.Fprintln(cmd.OutOrStdout())
				fmt.Fprintln(cmd.OutOrStdout(), `Use "kubecone explain CODE" for the recommended next step.`)
				return nil
			}
			code := strings.ToUpper(strings.TrimSpace(args[0]))
			info, ok := analyzer.LookupCode(code)
			if !ok {
				return fmt.Errorf("unknown code %q (try kubecone explain)", args[0])
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Code:    %s\n", info.Code)
			fmt.Fprintf(cmd.OutOrStdout(), "Title:   %s\n", info.Title)
			fmt.Fprintf(cmd.OutOrStdout(), "Summary: %s\n", info.Summary)
			fmt.Fprintf(cmd.OutOrStdout(), "Next:    %s\n", info.Next)
			return nil
		},
	}
	return cmd
}
