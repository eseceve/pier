package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <service>",
	Short: "Print the Kubernetes coordinates a service resolves to",
	Long: `Resolve a service name (and environment) to its concrete Kubernetes
coordinates without running anything, so agents and scripts can plan before
acting. Combine with --dry-run on other commands to preview the exact kubectl
invocation.`,
	Example: `  pier resolve api                 # resolved coordinates on the default env
  pier resolve api -e prod         # against the prod env
  pier resolve api -o json         # machine-readable, for agents/scripts`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}

		if flagOutput != "" {
			return encodeOutput(cmd.OutOrStdout(), flagOutput, target)
		}

		return renderTable(cmd.OutOrStdout(), [][]string{
			{"SERVICE", target.Service},
			{"ENV", target.Env},
			{"CONTEXT", target.Context},
			{"NAMESPACE", target.Namespace},
			{"DEPLOYMENT", target.Deployment},
			{"CONFIGMAP", target.ConfigMap},
			{"SECRET", target.Secret},
			{"PROTECTED", fmt.Sprintf("%t", target.Protected)},
		})
	},
}

func init() {
	addOutputFlag(resolveCmd)
	rootCmd.AddCommand(resolveCmd)
}
