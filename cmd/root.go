// Package cmd contains all pier subcommands.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/comparaonline/pier/internal/config"
	"github.com/comparaonline/pier/internal/confirm"
	"github.com/comparaonline/pier/internal/kube"
	"github.com/spf13/cobra"
)

var (
	flagEnv     string
	flagYes     bool
	flagDryRun  bool
	flagVerbose bool

	// runner is the kubectl execution seam; overridable in tests.
	runner kube.Runner = kube.ExecRunner{}
)

var rootCmd = &cobra.Command{
	Use:           "pier",
	Short:         "kubectl wrapper mapping service names to their k8s context/namespace",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// SetVersion injects build metadata from main.
func SetVersion(v, c, d string) {
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", v, c, d)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagEnv, "env", "e", "", "target environment (default: config defaultEnv)")
	rootCmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "skip confirmation for mutating ops")
	rootCmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "print the kubectl command without running it")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "print the kubectl command before running it")
}

// loadConfig loads the config from its default path.
func loadConfig() (*config.Config, error) {
	return config.Load(config.DefaultPath())
}

// resolve loads the config and resolves a service using the global --env flag.
func resolve(service string) (config.Target, error) {
	cfg, err := loadConfig()
	if err != nil {
		return config.Target{}, err
	}
	return cfg.Resolve(service, flagEnv)
}

// run honors --dry-run/--verbose, otherwise executes kubectl via the runner.
func run(args []string) error {
	if flagDryRun {
		fmt.Println("kubectl " + strings.Join(args, " "))
		return nil
	}
	if flagVerbose {
		fmt.Fprintln(os.Stderr, "+ kubectl "+strings.Join(args, " "))
	}
	return runner.Run(args)
}

// guard runs the confirmation prompt for a mutating op against a protected env,
// unless --yes was passed or --dry-run is active.
func guard(t config.Target, action string) error {
	if flagDryRun || flagYes || !t.Protected {
		return nil
	}
	return confirm.Prompt(rootCmd.InOrStdin(), rootCmd.ErrOrStderr(), t, action)
}

// joinArgs renders kubectl args for display.
func joinArgs(args []string) string { return strings.Join(args, " ") }
