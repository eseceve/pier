package cmd

import (
	"github.com/comparaonline/pier/internal/kube"
	"github.com/spf13/cobra"
)

var configGetCmd = &cobra.Command{
	Use:   "get <service>",
	Short: "Read the service's ConfigMap as YAML",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		return run(kube.ConfigGetArgs(target))
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
}
