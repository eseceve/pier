package cmd

import (
	"encoding/json"
	"errors"

	"github.com/eseceve/pier/internal/discover"
	"github.com/spf13/cobra"
)

var (
	discoverJSON      bool
	discoverContext   string
	discoverNamespace string

	// cmdRunner is the discovery seam; overridable in tests.
	cmdRunner discover.CmdRunner = discover.ExecRunner{}
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover cluster resources to help build the config",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if discoverNamespace != "" && discoverContext == "" {
			return errors.New("--namespace requires --context")
		}
		if !discoverJSON {
			_, err := cmd.OutOrStdout().Write([]byte(
				"Interactive discovery is coming in a later version.\n" +
					"For now, use `pier discover --json [--context X --namespace N]`,\n" +
					"or the pier-config skill to build your config conversationally.\n"))
			return err
		}

		out := map[string]any{}
		switch {
		case discoverContext == "":
			ctxs, err := discover.Contexts(cmdRunner)
			if err != nil {
				return err
			}
			out["contexts"] = ctxs
		case discoverNamespace == "":
			nss, err := discover.Namespaces(cmdRunner, discoverContext)
			if err != nil {
				return err
			}
			out["namespaces"] = nss
		default:
			deps, err := discover.Deployments(cmdRunner, discoverContext, discoverNamespace)
			if err != nil {
				return err
			}
			cms, err := discover.ConfigMaps(cmdRunner, discoverContext, discoverNamespace)
			if err != nil {
				return err
			}
			secs, err := discover.Secrets(cmdRunner, discoverContext, discoverNamespace)
			if err != nil {
				return err
			}
			out["deployments"] = deps
			out["configmaps"] = cms
			out["secrets"] = secs
		}

		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	},
}

func init() {
	discoverCmd.Flags().BoolVar(&discoverJSON, "json", false, "emit discovered data as JSON")
	discoverCmd.Flags().StringVar(&discoverContext, "context", "", "context to inspect")
	discoverCmd.Flags().StringVar(&discoverNamespace, "namespace", "", "namespace to inspect (requires --context)")
	rootCmd.AddCommand(discoverCmd)
}
