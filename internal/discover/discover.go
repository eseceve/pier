package discover

import "os/exec"

// CmdRunner runs a kubectl command and returns its captured stdout. It is the
// seam that tests replace.
type CmdRunner interface {
	Output(args []string) ([]byte, error)
}

// ExecRunner runs the real kubectl binary and captures stdout.
type ExecRunner struct{}

// Output runs `kubectl <args...>` and returns stdout.
func (ExecRunner) Output(args []string) ([]byte, error) {
	return exec.Command("kubectl", args...).Output() //nolint:gosec // G204: kubectl args are built internally by pier
}

// list runs a `kubectl ... -o name` query and returns the parsed names,
// optionally dropping Kubernetes-managed noise.
func list(r CmdRunner, filter bool, args ...string) ([]string, error) {
	out, err := r.Output(args)
	if err != nil {
		return nil, err
	}
	names := parseNames(out)
	if filter {
		names = filterNoise(names)
	}
	return names, nil
}

// Contexts lists the kubeconfig context names.
func Contexts(r CmdRunner) ([]string, error) {
	return list(r, false, "config", "get-contexts", "-o", "name")
}

// Namespaces lists the namespaces in a context.
func Namespaces(r CmdRunner, ctx string) ([]string, error) {
	return list(r, false, "--context", ctx, "get", "namespaces", "-o", "name")
}

// Deployments lists deployment names in a (context, namespace).
func Deployments(r CmdRunner, ctx, ns string) ([]string, error) {
	return list(r, false, "--context", ctx, "--namespace", ns, "get", "deployments", "-o", "name")
}

// ConfigMaps lists ConfigMap names in a (context, namespace), filtering noise.
func ConfigMaps(r CmdRunner, ctx, ns string) ([]string, error) {
	return list(r, true, "--context", ctx, "--namespace", ns, "get", "configmaps", "-o", "name")
}

// Secrets lists Secret names in a (context, namespace), filtering noise.
func Secrets(r CmdRunner, ctx, ns string) ([]string, error) {
	return list(r, true, "--context", ctx, "--namespace", ns, "get", "secrets", "-o", "name")
}
