// Package kube builds and runs kubectl commands from a resolved Target.
package kube

import (
	"strconv"

	"github.com/comparaonline/pier/internal/config"
)

func base(t config.Target, args ...string) []string {
	return append([]string{"--context", t.Context, "--namespace", t.Namespace}, args...)
}

// RestartArgs restarts the service's deployment.
func RestartArgs(t config.Target) []string {
	return base(t, "rollout", "restart", "deployment/"+t.Deployment)
}

// RolloutStatusArgs reports the deployment's rollout status.
func RolloutStatusArgs(t config.Target) []string {
	return base(t, "rollout", "status", "deployment/"+t.Deployment)
}

// GetDeploymentArgs shows the deployment summary.
func GetDeploymentArgs(t config.Target) []string {
	return base(t, "get", "deployment/"+t.Deployment)
}

// LogsArgs streams the deployment's logs.
func LogsArgs(t config.Target, follow bool, tail int, container string) []string {
	args := base(t, "logs", "deployment/"+t.Deployment)
	if follow {
		args = append(args, "-f")
	}
	if tail > 0 {
		args = append(args, "--tail="+strconv.Itoa(tail))
	}
	if container != "" {
		args = append(args, "-c", container)
	}
	return args
}

// ConfigGetArgs reads the service's ConfigMap as YAML.
func ConfigGetArgs(t config.Target) []string {
	return base(t, "get", "configmap/"+t.ConfigMap, "-o", "yaml")
}

// ConfigEditArgs opens the service's ConfigMap in $EDITOR.
func ConfigEditArgs(t config.Target) []string {
	return base(t, "edit", "configmap/"+t.ConfigMap)
}

// SecretGetArgs reads the service's Secret. When decode is true, values are
// base64-decoded via a go-template.
func SecretGetArgs(t config.Target, decode bool) []string {
	if decode {
		tmpl := `go-template={{range $k,$v := .data}}{{$k}}: {{$v | base64decode}}{{"\n"}}{{end}}`
		return base(t, "get", "secret/"+t.Secret, "-o", tmpl)
	}
	return base(t, "get", "secret/"+t.Secret, "-o", "yaml")
}

// SecretEditArgs opens the service's Secret in $EDITOR.
func SecretEditArgs(t config.Target) []string {
	return base(t, "edit", "secret/"+t.Secret)
}
