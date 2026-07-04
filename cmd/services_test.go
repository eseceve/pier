package cmd

import (
	"bytes"
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
