package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("defaultEnv: staging\nenvironments:\n  staging:\n    context: staging-cluster\nservices:\n  api:\n    namespace: default\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultEnv != "staging" || cfg.Environments["staging"].Context != "staging-cluster" {
		t.Fatalf("parsed wrong: %+v", cfg)
	}
}

func TestLoadMissingFileErrors(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDefaultPathHonorsEnvVar(t *testing.T) {
	t.Setenv("PIER_CONFIG", "/custom/pier.yaml")
	if got := DefaultPath(); got != "/custom/pier.yaml" {
		t.Fatalf("got %q", got)
	}
}
