package cmd

import (
	"bytes"
	"testing"
)

// fakeRunner records the args it was asked to run.
type fakeRunner struct{ got []string }

func (f *fakeRunner) Run(args []string) error { f.got = args; return nil }

// withConfig points PIER_CONFIG at a temp config and resets flags.
func withConfig(t *testing.T, yaml string) {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/config.yaml"
	if err := writeFile(path, yaml); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PIER_CONFIG", path)
	flagEnv, flagYes, flagDryRun, flagVerbose = "", false, false, false
}

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

func TestRestartRunsRolloutRestart(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	defer func() { runner = nil }()

	rootCmd.SetArgs([]string{"restart", "api"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "--context staging-cluster --namespace default rollout restart deployment/api"
	if joinArgs(fake.got) != want {
		t.Fatalf("got %q want %q", joinArgs(fake.got), want)
	}
}

func TestRestartProtectedEnvBlockedWithoutConfirmation(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	defer func() { runner = nil }()
	flagEnv = "prod" // protected

	rootCmd.SetArgs([]string{"restart", "api", "-e", "prod"})
	rootCmd.SetIn(bytes.NewReader([]byte("wrong\n")))
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected confirmation to block the restart")
	}
	if fake.got != nil {
		t.Fatalf("runner should not have been called, got %v", fake.got)
	}
}
