package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestSecretGetPlain(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake

	rootCmd.SetArgs([]string{"secret", "get", "api"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "--context staging-cluster --namespace default get secret/api -o yaml"
	if joinArgs(fake.got) != want {
		t.Fatalf("got %q want %q", joinArgs(fake.got), want)
	}
}

func TestSecretGetDecode(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake

	rootCmd.SetArgs([]string{"secret", "get", "api", "--decode"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(joinArgs(fake.got), "base64decode") {
		t.Fatalf("decode not applied: %q", joinArgs(fake.got))
	}
}

func TestSecretEditFailsFastNonInteractive(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	isTTY = func(io.Reader) bool { return false }

	rootCmd.SetArgs([]string{"secret", "edit", "api"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected secret edit to fail fast without a TTY")
	}
	if fake.got != nil {
		t.Fatalf("runner should not have been called, got %v", fake.got)
	}
}
