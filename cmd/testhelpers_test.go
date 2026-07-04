package cmd

import (
	"os"
	"testing"
)

// fakeRunner records the args it was asked to run; shared by command tests.
type fakeRunner struct{ got []string }

func (f *fakeRunner) Run(args []string) error { f.got = args; return nil }

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o600)
}

// sampleYAML is a minimal config used across command tests.
const sampleYAML = `defaultEnv: staging
environments:
  staging:
    context: staging-cluster
  prod:
    context: prod-cluster
    protected: true
services:
  api:
    namespace: default
`

// withConfig points PIER_CONFIG at a temp config, resets flags, and restores the
// runner seam after the test.
func withConfig(t *testing.T, yaml string) {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/config.yaml"
	if err := writeFile(path, yaml); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PIER_CONFIG", path)
	flagEnv, flagYes, flagDryRun, flagVerbose = "", false, false, false
	prev := runner
	t.Cleanup(func() { runner = prev })
}
