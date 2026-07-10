package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "List configured services with their namespace and deployment",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintln(w, "SERVICE\tNAMESPACE\tDEPLOYMENT"); err != nil {
			return err
		}
		for _, name := range cfg.ServiceNames() {
			svc := cfg.Services[name]
			deployment := svc.Deployment
			if deployment == "" {
				deployment = name
			}
			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", name, svc.Namespace, deployment); err != nil {
				return err
			}
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(servicesCmd)
}
