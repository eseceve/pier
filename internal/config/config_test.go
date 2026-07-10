package config

import "testing"

func sampleConfig() *Config {
	return &Config{
		DefaultEnv: "staging",
		Environments: map[string]Environment{
			"staging": {Context: "staging-cluster"},
			"prod":    {Context: "prod-cluster", Protected: true},
		},
		Services: map[string]Service{
			"api": {Namespace: "default"},
			"web": {
				Namespace:  "default",
				Deployment: "web-server",
				Overrides:  map[string]ServiceOverride{"prod": {Namespace: "web-prod"}},
			},
		},
	}
}

func TestResolveDefaultsToDefaultEnv(t *testing.T) {
	got, err := sampleConfig().Resolve("api", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Env != "staging" || got.Context != "staging-cluster" {
		t.Fatalf("got %+v", got)
	}
	if got.Namespace != "default" || got.Deployment != "api" {
		t.Fatalf("deployment/namespace default wrong: %+v", got)
	}
	if got.Protected {
		t.Fatalf("staging should not be protected")
	}
}

func TestResolveDeploymentOverrideAndNamespacePerEnv(t *testing.T) {
	got, err := sampleConfig().Resolve("web", "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Deployment != "web-server" {
		t.Fatalf("deployment override not applied: %+v", got)
	}
	if got.Namespace != "web-prod" {
		t.Fatalf("per-env namespace override not applied: %+v", got)
	}
	if !got.Protected {
		t.Fatalf("prod should be protected")
	}
}

func TestResolveConfigMapAndSecretOverridePerEnv(t *testing.T) {
	cfg := &Config{
		DefaultEnv:   "staging",
		Environments: map[string]Environment{"prod": {Context: "prod-cluster", Protected: true}},
		Services: map[string]Service{
			"web": {
				Namespace: "default",
				Overrides: map[string]ServiceOverride{
					"prod": {ConfigMap: "web-prod-cm", Secret: "web-prod-secret"},
				},
			},
		},
	}
	got, err := cfg.Resolve("web", "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ConfigMap != "web-prod-cm" {
		t.Fatalf("per-env configmap override not applied: %+v", got)
	}
	if got.Secret != "web-prod-secret" {
		t.Fatalf("per-env secret override not applied: %+v", got)
	}
}

func TestResolveUnknownServiceErrors(t *testing.T) {
	_, err := sampleConfig().Resolve("nope", "staging")
	if err == nil {
		t.Fatal("expected error for unknown service")
	}
}

func TestResolveUnknownEnvErrors(t *testing.T) {
	_, err := sampleConfig().Resolve("api", "qa")
	if err == nil {
		t.Fatal("expected error for unknown environment")
	}
}
