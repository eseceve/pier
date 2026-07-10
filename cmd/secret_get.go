package cmd

import (
	"github.com/comparaonline/pier/internal/kube"
	"github.com/spf13/cobra"
)

var secretDecode bool

var secretGetCmd = &cobra.Command{
	Use:   "get <service>",
	Short: "Read the service's Secret (use --decode for cleartext)",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		return run(kube.SecretGetArgs(target, secretDecode))
	},
}

func init() {
	secretGetCmd.Flags().BoolVar(&secretDecode, "decode", false, "print secret values in cleartext")
	secretCmd.AddCommand(secretGetCmd)
}
