package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestConfigGetReadsConfigMap(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake

	rootCmd.SetArgs([]string{"config", "get", "api"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "--context staging-cluster --namespace default get configmap/api -o yaml"
	if joinArgs(fake.got) != want {
		t.Fatalf("got %q want %q", joinArgs(fake.got), want)
	}
}

func TestConfigEditFailsFastNonInteractive(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	isTTY = func(io.Reader) bool { return false }

	rootCmd.SetArgs([]string{"config", "edit", "api"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected config edit to fail fast without a TTY")
	}
	if fake.got != nil {
		t.Fatalf("runner should not have been called, got %v", fake.got)
	}
}

func TestConfigEditDryRunSkipsTTYRequirement(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	isTTY = func(io.Reader) bool { return false }

	rootCmd.SetArgs([]string{"config", "edit", "api", "--dry-run"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("dry-run should not require a TTY, got: %v", err)
	}
	if !strings.Contains(out.String(), "edit configmap/api") {
		t.Fatalf("dry-run should print the kubectl command, got: %q", out.String())
	}
}

func TestConfigInitWritesTemplate(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.yaml"
	t.Setenv("PIER_CONFIG", path)
	flagEnv, flagYes, flagDryRun, flagVerbose, flagOutput = "", false, false, false, ""

	rootCmd.SetArgs([]string{"config", "init"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path) //nolint:gosec // G304: test reads a temp file it just wrote
	if err != nil {
		t.Fatalf("config not written: %v", err)
	}
	if !strings.Contains(string(data), "defaultEnv:") {
		t.Fatalf("template missing expected content: %q", string(data))
	}
}
