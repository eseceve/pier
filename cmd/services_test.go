package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestServicesListsConfiguredServices(t *testing.T) {
	withConfig(t, sampleYAML)

	rootCmd.SetArgs([]string{"services"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "api") {
		t.Fatalf("expected 'api' in output: %q", out.String())
	}
}

func TestServicesJSONListsConfiguredServices(t *testing.T) {
	withConfig(t, sampleYAML)

	rootCmd.SetArgs([]string{"services", "-o", "json"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	if len(got) != 1 || got[0]["name"] != "api" || got[0]["namespace"] != "default" {
		t.Fatalf("unexpected services JSON: %v", got)
	}
}
