// Package config loads pier's YAML config and resolves a (service, env) pair
// into the concrete Kubernetes coordinates a command needs.
package config

import (
	"fmt"
	"sort"
	"strings"
)

// Environment is a named Kubernetes context, optionally protected.
type Environment struct {
	Context   string `yaml:"context"`
	Protected bool   `yaml:"protected"`
}

// ServiceOverride overrides a service's fields for a specific environment.
type ServiceOverride struct {
	Namespace string `yaml:"namespace"`
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

	namespace := svc.Namespace
	if o, ok := svc.Overrides[env]; ok && o.Namespace != "" {
		namespace = o.Namespace
	}

	return Target{
		Service:    service,
		Env:        env,
		Context:    environment.Context,
		Namespace:  namespace,
		Deployment: orDefault(svc.Deployment, service),
		ConfigMap:  orDefault(svc.ConfigMap, service),
		Secret:     orDefault(svc.Secret, service),
		Protected:  environment.Protected,
	}, nil
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
