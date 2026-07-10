package cmd

import (
	"github.com/spf13/cobra"
)

// serviceRow is the structured shape of one configured service.
type serviceRow struct {
	Name       string `json:"name" yaml:"name"`
	Namespace  string `json:"namespace" yaml:"namespace"`
	Deployment string `json:"deployment" yaml:"deployment"`
}

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "List configured services with their namespace and deployment",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		rows := make([]serviceRow, 0, len(cfg.Services))
		for _, name := range cfg.ServiceNames() {
			svc := cfg.Services[name]
			deployment := svc.Deployment
			if deployment == "" {
				deployment = name
			}
			rows = append(rows, serviceRow{Name: name, Namespace: svc.Namespace, Deployment: deployment})
		}

		if flagOutput != "" {
			return encodeOutput(cmd.OutOrStdout(), flagOutput, rows)
		}

		table := make([][]string, 0, len(rows)+1)
		table = append(table, []string{"SERVICE", "NAMESPACE", "DEPLOYMENT"})
		for _, r := range rows {
			table = append(table, []string{r.Name, r.Namespace, r.Deployment})
		}
		return renderTable(cmd.OutOrStdout(), table)
	},
}

func init() {
	addOutputFlag(servicesCmd)
	rootCmd.AddCommand(servicesCmd)
}
