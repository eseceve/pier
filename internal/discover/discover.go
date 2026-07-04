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

// Contexts lists the kubeconfig context names.
func Contexts(r CmdRunner) ([]string, error) {
	out, err := r.Output([]string{"config", "get-contexts", "-o", "name"})
	if err != nil {
		return nil, err
	}
	return parseNames(out), nil
}

// Namespaces lists the namespaces in a context.
func Namespaces(r CmdRunner, ctx string) ([]string, error) {
	out, err := r.Output([]string{"--context", ctx, "get", "namespaces", "-o", "name"})
	if err != nil {
		return nil, err
	}
	return parseNames(out), nil
}

// Deployments lists deployment names in a (context, namespace).
func Deployments(r CmdRunner, ctx, ns string) ([]string, error) {
	out, err := r.Output([]string{"--context", ctx, "--namespace", ns, "get", "deployments", "-o", "name"})
	if err != nil {
		return nil, err
	}
	return parseNames(out), nil
}

// ConfigMaps lists ConfigMap names in a (context, namespace), filtering noise.
func ConfigMaps(r CmdRunner, ctx, ns string) ([]string, error) {
	out, err := r.Output([]string{"--context", ctx, "--namespace", ns, "get", "configmaps", "-o", "name"})
	if err != nil {
		return nil, err
	}
	return filterNoise(parseNames(out)), nil
}

// Secrets lists Secret names in a (context, namespace), filtering noise.
func Secrets(r CmdRunner, ctx, ns string) ([]string, error) {
	out, err := r.Output([]string{"--context", ctx, "--namespace", ns, "get", "secrets", "-o", "name"})
	if err != nil {
		return nil, err
	}
	return filterNoise(parseNames(out)), nil
}
