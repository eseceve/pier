package cmd

import (
	"github.com/comparaonline/pier/internal/kube"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <service>",
	Short: "Show the service's deployment and rollout status",
	Long: `Show the service's deployment summary and rollout status. For structured
resource data, append kubectl's own -o (e.g. use kubectl directly), since status
forwards kubectl's output verbatim.`,
	Example: `  pier status api                  # pods + rollout status on staging
  pier status api -e prod`,
	Args: cobra.ExactArgs(1),
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
