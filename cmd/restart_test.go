package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestRestartRunsRolloutRestart(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake

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
	isTTY = func(io.Reader) bool { return true }

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

func TestRestartProtectedEnvFailsFastNonInteractive(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	isTTY = func(io.Reader) bool { return false }

	rootCmd.SetArgs([]string{"restart", "api", "-e", "prod"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected a fail-fast error against a protected env without a TTY")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("error should point the caller at --yes, got: %v", err)
	}
	if fake.got != nil {
		t.Fatalf("runner should not have been called, got %v", fake.got)
	}
}
