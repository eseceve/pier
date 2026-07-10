package cmd

import (
	"bytes"
	"testing"
)

// recordingRunner records every invocation (status runs two).
type recordingRunner struct{ calls [][]string }

func (r *recordingRunner) Run(args []string) error {
	r.calls = append(r.calls, args)
	return nil
}

func TestStatusRunsGetAndRolloutStatus(t *testing.T) {
	withConfig(t, sampleYAML)
	rec := &recordingRunner{}
	runner = rec

	rootCmd.SetArgs([]string{"status", "api"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rec.calls) != 2 {
		t.Fatalf("expected 2 kubectl calls, got %d", len(rec.calls))
	}
	if joinArgs(rec.calls[0]) != "--context staging-cluster --namespace default get deployment/api" {
		t.Fatalf("call 0 wrong: %q", joinArgs(rec.calls[0]))
	}
	if joinArgs(rec.calls[1]) != "--context staging-cluster --namespace default rollout status deployment/api" {
		t.Fatalf("call 1 wrong: %q", joinArgs(rec.calls[1]))
	}
}

func TestLogsPassesFlags(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake

	rootCmd.SetArgs([]string{"logs", "api", "-f", "--tail", "50"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "--context staging-cluster --namespace default logs deployment/api -f --tail=50"
	if joinArgs(fake.got) != want {
		t.Fatalf("got %q want %q", joinArgs(fake.got), want)
	}
}
