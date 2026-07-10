package discover

import (
	"reflect"
	"testing"
)

func TestParseNamesStripsKindPrefix(t *testing.T) {
	got := parseNames([]byte("deployment.apps/api\ndeployment.apps/web\n\n"))
	want := []string{"api", "web"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseNamesHandlesNoPrefix(t *testing.T) {
	got := parseNames([]byte("staging-ctx\nprod-ctx\n"))
	want := []string{"staging-ctx", "prod-ctx"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestFilterNoiseDropsHelmAndTokens(t *testing.T) {
	in := []string{"api-local", "sh.helm.release.v1.api.v1", "default-token-abcde", "kube-root-ca.crt", "istio-ca-root-cert"}
	got := filterNoise(in)
	want := []string{"api-local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMatchExact(t *testing.T) {
	if got := Match("api", []string{"api", "other"}); got != "api" {
		t.Fatalf("got %q", got)
	}
}

func TestMatchUniquePrefix(t *testing.T) {
	got := Match("ai-interactor", []string{"ai-interactor-local", "mcp-ai-interactor-local"})
	if got != "ai-interactor-local" {
		t.Fatalf("got %q", got)
	}
}

func TestMatchAmbiguousReturnsEmpty(t *testing.T) {
	if got := Match("x", []string{"x-a", "x-b"}); got != "" {
		t.Fatalf("expected empty on ambiguous, got %q", got)
	}
}

func TestMatchNoneReturnsEmpty(t *testing.T) {
	if got := Match("api", []string{"foo", "bar"}); got != "" {
		t.Fatalf("got %q", got)
	}
}
