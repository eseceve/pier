package cmd

import (
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
)

var configGetCmd = &cobra.Command{
	Use:     "get <service>",
	Short:   "Read the service's ConfigMap as YAML",
	Long:    "Read the service's ConfigMap. Output is kubectl's YAML, forwarded verbatim.",
	Example: "  pier config get api",
	Args:    cobra.ExactArgs(1),
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
