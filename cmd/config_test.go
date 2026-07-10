package cmd

import (
	"bytes"
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

func TestConfigInitWritesTemplate(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.yaml"
	t.Setenv("PIER_CONFIG", path)
	flagEnv, flagYes, flagDryRun, flagVerbose = "", false, false, false

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
