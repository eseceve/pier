package cmd

import (
	"github.com/comparaonline/pier/internal/kube"
	"github.com/spf13/cobra"
)

var configEditCmd = &cobra.Command{
	Use:   "edit <service>",
	Short: "Edit the service's ConfigMap in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		kubectlArgs := kube.ConfigEditArgs(target)
		if err := guard(target, kubectlCmd(kubectlArgs)); err != nil {
			return err
		}
		return run(kubectlArgs)
	},
}

func init() {
	configCmd.AddCommand(configEditCmd)
}
