// Package cmd contains all pier subcommands.
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/eseceve/pier/internal/config"
	"github.com/eseceve/pier/internal/confirm"
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	flagEnv     string
	flagYes     bool
	flagDryRun  bool
	flagVerbose bool
	flagOutput  string

	// runner is the kubectl execution seam; overridable in tests.
	runner kube.Runner = kube.ExecRunner{}

	// isTTY reports whether the reader is an interactive terminal; the seam lets
	// tests exercise the prompt and fail-fast paths without a real TTY.
	isTTY = defaultIsTTY
)

// defaultIsTTY treats a reader as interactive only when it is an *os.File backed
// by a character device (a terminal), so pipes and non-TTY callers fail fast.
func defaultIsTTY(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

var rootCmd = &cobra.Command{
	Use:   "pier",
	Short: "kubectl wrapper mapping service names to their k8s context/namespace",
	Long: `pier maps a friendly service name to its Kubernetes coordinates
(context + namespace + deployment/configmap/secret) via a config file, so
day-to-day operations are short and memorable — targeting staging by default.

Agent/script-friendly contract:
  - --dry-run prints the exact kubectl command without running it.
  - pier resolve -o json reports a service's resolved coordinates (no side effects).
  - services and resolve accept -o json|yaml; pass-through commands (logs,
    status, config get, secret get) delegate structured output to kubectl's own -o.
  - Structured output goes to stdout; diagnostics and prompts go to stderr.
  - Mutating ops against a protected env require a TTY or an explicit --yes;
    non-interactive callers fail fast instead of hanging.`,
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

// Root returns the root command. It exists so the docs generator can walk the
// command tree; application code should call Execute instead.
func Root() *cobra.Command {
	return rootCmd
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
		_, err := fmt.Fprintln(rootCmd.OutOrStdout(), kubectlCmd(args))
		return err
	}
	if flagVerbose {
		if _, err := fmt.Fprintln(rootCmd.ErrOrStderr(), "+ "+kubectlCmd(args)); err != nil {
			return err
		}
	}
	return runner.Run(args)
}

// requireTTY returns an error when stdin is not an interactive terminal, so
// $EDITOR-backed operations fail fast for non-interactive callers instead of
// hanging on a blocking editor.
func requireTTY(what string) error {
	if isTTY(rootCmd.InOrStdin()) {
		return nil
	}
	return fmt.Errorf("%s requires an interactive terminal; not available non-interactively", what)
}

// guard runs the confirmation prompt for a mutating op against a protected env,
// unless --yes was passed or --dry-run is active.
func guard(t config.Target, action string) error {
	if flagDryRun || flagYes || !t.Protected {
		return nil
	}
	if !isTTY(rootCmd.InOrStdin()) {
		return fmt.Errorf("refusing to run a mutating op against protected env %q non-interactively; pass --yes to confirm", t.Env)
	}
	return confirm.Prompt(rootCmd.InOrStdin(), rootCmd.ErrOrStderr(), t, action)
}

// runInteractive runs an $EDITOR-backed op: it requires a TTY (skipped under
// --dry-run), applies the protected-env guard, then runs the command.
func runInteractive(t config.Target, args []string, what string) error {
	if !flagDryRun {
		if err := requireTTY(what); err != nil {
			return err
		}
	}
	if err := guard(t, kubectlCmd(args)); err != nil {
		return err
	}
	return run(args)
}

// addOutputFlag registers the -o/--output flag on a pier-owned command whose
// data pier fully controls (json or yaml). Pass-through commands delegate
// structured output to kubectl's own -o instead.
func addOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&flagOutput, "output", "o", "", "structured output format: json or yaml")
}

// encodeOutput writes v to w in the given structured format (json or yaml).
func encodeOutput(w io.Writer, format string, v any) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	case "yaml":
		enc := yaml.NewEncoder(w)
		enc.SetIndent(2)
		if err := enc.Encode(v); err != nil {
			return err
		}
		return enc.Close()
	default:
		return fmt.Errorf("unknown output format %q (want json or yaml)", format)
	}
}

// renderTable writes rows as a padded, tab-aligned table.
func renderTable(w io.Writer, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, row := range rows {
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// joinArgs renders kubectl args for display.
func joinArgs(args []string) string { return strings.Join(args, " ") }

// kubectlCmd renders kubectl args as a display string, e.g. "kubectl get pods".
func kubectlCmd(args []string) string { return "kubectl " + joinArgs(args) }
