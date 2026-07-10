package cmd

import (
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
)

var (
	logsFollow    bool
	logsTail      int
	logsContainer string
)

var logsCmd = &cobra.Command{
	Use:   "logs <service>",
	Short: "Show the service's deployment logs",
	Example: `  pier logs api -f                 # follow logs on staging
  pier logs api --tail 100 -c app  # last 100 lines from the app container`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		return run(kube.LogsArgs(target, logsFollow, logsTail, logsContainer))
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "stream logs")
	logsCmd.Flags().IntVar(&logsTail, "tail", 0, "lines of recent log to show")
	logsCmd.Flags().StringVarP(&logsContainer, "container", "c", "", "container name")
	rootCmd.AddCommand(logsCmd)
}
