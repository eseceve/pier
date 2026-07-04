package cmd

import (
	"github.com/comparaonline/pier/internal/kube"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <service>",
	Short: "Show the service's deployment and rollout status",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		if err := run(kube.GetDeploymentArgs(target)); err != nil {
			return err
		}
		return run(kube.RolloutStatusArgs(target))
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
