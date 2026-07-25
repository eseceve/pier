package confirm

import (
	"bytes"
	"strings"
	"testing"

	"github.com/eseceve/pier/internal/config"
)

func protectedTarget() config.Target {
	return config.Target{
		Service: "api", Env: "prod",
		Context: "prod-cluster", Namespace: "default", Protected: true,
	}
}

func TestPromptAcceptsMatchingServiceName(t *testing.T) {
	var out bytes.Buffer
	err := Prompt(strings.NewReader("api\n"), &out, protectedTarget(), "kubectl rollout restart deployment/api")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !strings.Contains(out.String(), "PROD") || !strings.Contains(out.String(), "prod-cluster") {
		t.Fatalf("prompt is not env-forward: %q", out.String())
	}
}

func TestPromptRejectsWrongInput(t *testing.T) {
	var out bytes.Buffer
	err := Prompt(strings.NewReader("nope\n"), &out, protectedTarget(), "action")
	if err == nil {
		t.Fatal("expected error on mismatch")
	}
}
