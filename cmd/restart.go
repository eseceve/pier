package cmd

import (
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart <service>",
	Short: "Rollout restart of the service's deployment",
	Example: `  pier restart api                 # rollout restart on staging
  pier restart api -e prod         # prompts for confirmation (or --yes)
  pier restart api -e prod --dry-run   # print the kubectl command only`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		kubectlArgs := kube.RestartArgs(target)
		if err := guard(target, kubectlCmd(kubectlArgs)); err != nil {
			return err
		}
		return run(kubectlArgs)
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
