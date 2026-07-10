package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eseceve/pier/internal/config"
	"github.com/spf13/cobra"
)

const configTemplate = `defaultEnv: staging

environments:
  staging:
    context: staging-cluster
  prod:
    context: prod-cluster
    protected: true

services:
  api:
    namespace: default
  web:
    namespace: default
    deployment: web-server
    overrides:
      prod:
        namespace: web-prod
`

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Write a starter config file",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		path := config.DefaultPath()
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config already exists at %s", path)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil { //nolint:gosec // G301: user config dir, 750 is intentional
			return err
		}
		if err := os.WriteFile(path, []byte(configTemplate), 0o600); err != nil { //nolint:gosec // G306: 600 is intentional for config file
			return err
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", path)
		return err
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
}
