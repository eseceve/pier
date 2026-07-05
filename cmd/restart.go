package cmd

import (
	"github.com/comparaonline/pier/internal/kube"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart <service>",
	Short: "Rollout restart of the service's deployment",
	Args:  cobra.ExactArgs(1),
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
