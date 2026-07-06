package cmd

import (
	"github.com/comparaonline/pier/internal/kube"
	"github.com/spf13/cobra"
)

var configEditCmd = &cobra.Command{
	Use:   "edit <service>",
	Short: "Edit the service's ConfigMap in $EDITOR",
	Long:  "Edit the service's ConfigMap in $EDITOR. Interactive-only: requires a TTY and fails fast for non-interactive callers.",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		return runInteractive(target, kube.ConfigEditArgs(target), "config edit")
	},
}

func init() {
	configCmd.AddCommand(configEditCmd)
}
