package kube

import (
	"fmt"
	"os"
	"os/exec"
)

// Runner executes a kubectl invocation. It is the seam that tests replace.
type Runner interface {
	Run(args []string) error
}

// ExecRunner runs the real kubectl binary, inheriting the process stdio (so
// interactive commands like `edit` and streaming `logs -f` work).
type ExecRunner struct{}

// Run executes `kubectl <args...>`. It returns *exec.ExitError on non-zero exit
// so the caller can propagate kubectl's exit code.
func (ExecRunner) Run(args []string) error {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl not found in PATH")
	}
	cmd := exec.Command("kubectl", args...) //nolint:gosec // G204: kubectl args are built internally by pier
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
