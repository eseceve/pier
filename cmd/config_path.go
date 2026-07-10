package cmd

import (
	"fmt"

	"github.com/eseceve/pier/internal/config"
	"github.com/spf13/cobra"
)

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the resolved config file path",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), config.DefaultPath())
		return err
	},
}

func init() {
	configCmd.AddCommand(configPathCmd)
}
