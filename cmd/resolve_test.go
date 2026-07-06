package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestResolveJSONReportsCoordinates(t *testing.T) {
	withConfig(t, sampleYAML)

	rootCmd.SetArgs([]string{"resolve", "api", "-o", "json"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	want := map[string]string{
		"service":    "api",
		"env":        "staging",
		"context":    "staging-cluster",
		"namespace":  "default",
		"deployment": "api",
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("field %q = %v, want %q", k, got[k], v)
		}
	}
	if got["protected"] != false {
		t.Fatalf("staging should not be protected: %v", got["protected"])
	}
}

func TestResolveHumanReadable(t *testing.T) {
	withConfig(t, sampleYAML)

	rootCmd.SetArgs([]string{"resolve", "api"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "staging-cluster") || !strings.Contains(out.String(), "api") {
		t.Fatalf("human output missing coordinates: %q", out.String())
	}
}

func TestResolveUnknownFormatErrors(t *testing.T) {
	withConfig(t, sampleYAML)

	rootCmd.SetArgs([]string{"resolve", "api", "-o", "toml"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected an error for an unsupported output format")
	}
}
