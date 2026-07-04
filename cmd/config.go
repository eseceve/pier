package cmd

import "github.com/spf13/cobra"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Read, edit, locate or create the config / a service's ConfigMap",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
