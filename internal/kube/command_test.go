package kube

import (
	"strings"
	"testing"

	"github.com/eseceve/pier/internal/config"
)

func target() config.Target {
	return config.Target{
		Service: "api", Env: "staging",
		Context: "staging-cluster", Namespace: "default",
		Deployment: "api", ConfigMap: "api", Secret: "api",
	}
}

func joined(args []string) string { return strings.Join(args, " ") }

func TestRestartArgs(t *testing.T) {
	got := joined(RestartArgs(target()))
	want := "--context staging-cluster --namespace default rollout restart deployment/api"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLogsArgsWithFlags(t *testing.T) {
	got := joined(LogsArgs(target(), true, 100, "app"))
	want := "--context staging-cluster --namespace default logs deployment/api -f --tail=100 -c app"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLogsArgsDefaults(t *testing.T) {
	got := joined(LogsArgs(target(), false, 0, ""))
	want := "--context staging-cluster --namespace default logs deployment/api"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSecretGetArgsDecode(t *testing.T) {
	got := SecretGetArgs(target(), true)
	last := got[len(got)-1]
	if !strings.Contains(last, "base64decode") {
		t.Fatalf("decode template missing: %q", last)
	}
}

func TestConfigGetArgs(t *testing.T) {
	got := joined(ConfigGetArgs(target()))
	want := "--context staging-cluster --namespace default get configmap/api -o yaml"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
