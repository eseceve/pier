package cmd

import "github.com/spf13/cobra"

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Read or edit a service's Secret",
}

func init() {
	rootCmd.AddCommand(secretCmd)
}
