// Package config loads pier's YAML config and resolves a (service, env) pair
// into the concrete Kubernetes coordinates a command needs.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Environment is a named Kubernetes context, optionally protected.
type Environment struct {
	Context   string `yaml:"context"`
	Protected bool   `yaml:"protected"`
}

// ServiceOverride overrides a service's fields for a specific environment.
type ServiceOverride struct {
	Namespace string `yaml:"namespace"`
	ConfigMap string `yaml:"configmap"`
	Secret    string `yaml:"secret"`
}

// Service maps a friendly name to its Kubernetes resources. Deployment,
// ConfigMap and Secret default to the service name when empty.
type Service struct {
	Namespace  string                     `yaml:"namespace"`
	Deployment string                     `yaml:"deployment"`
	ConfigMap  string                     `yaml:"configmap"`
	Secret     string                     `yaml:"secret"`
	Overrides  map[string]ServiceOverride `yaml:"overrides"`
}

// Config is the whole pier configuration file.
type Config struct {
	DefaultEnv   string                 `yaml:"defaultEnv"`
	Environments map[string]Environment `yaml:"environments"`
	Services     map[string]Service     `yaml:"services"`
}

// Target is a fully resolved command destination.
type Target struct {
	Service    string
	Env        string
	Context    string
	Namespace  string
	Deployment string
	ConfigMap  string
	Secret     string
	Protected  bool
}

// ServiceNames returns the configured service names, sorted.
func (c *Config) ServiceNames() []string {
	names := make([]string, 0, len(c.Services))
	for name := range c.Services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Resolve turns a (service, env) pair into a Target. An empty env falls back to
// DefaultEnv.
func (c *Config) Resolve(service, env string) (Target, error) {
	if env == "" {
		env = c.DefaultEnv
	}

	environment, ok := c.Environments[env]
	if !ok {
		return Target{}, fmt.Errorf("unknown environment %q", env)
	}

	svc, ok := c.Services[service]
	if !ok {
		return Target{}, fmt.Errorf("unknown service %q (configured: %s)", service, strings.Join(c.ServiceNames(), ", "))
	}

	override := svc.Overrides[env]

	return Target{
		Service:    service,
		Env:        env,
		Context:    environment.Context,
		Namespace:  orDefault(override.Namespace, svc.Namespace),
		Deployment: orDefault(svc.Deployment, service),
		ConfigMap:  orDefault(override.ConfigMap, orDefault(svc.ConfigMap, service)),
		Secret:     orDefault(override.Secret, orDefault(svc.Secret, service)),
		Protected:  environment.Protected,
	}, nil
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// DefaultPath returns the config path: $PIER_CONFIG if set, else
// $XDG_CONFIG_HOME/pier/config.yaml, else ~/.config/pier/config.yaml.
func DefaultPath() string {
	if p := os.Getenv("PIER_CONFIG"); p != "" {
		return p
	}
	if base := os.Getenv("XDG_CONFIG_HOME"); base != "" {
		return filepath.Join(base, "pier", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pier", "config.yaml")
}

// Load reads and parses the config file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: config path is user-provided by design
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return &cfg, nil
}
