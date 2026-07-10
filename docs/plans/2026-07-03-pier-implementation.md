# pier Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `pier`, a Go CLI that wraps `kubectl` and maps a friendly service name to its Kubernetes coordinates (context + namespace + resource names) via a YAML config, defaulting to staging.

**Architecture:** Pure, testable core packages (`internal/config` resolution, `internal/kube` command building + a `Runner` seam, `internal/confirm` guardrail) wired by a thin cobra command layer (`cmd/`). `pier` shells out to `kubectl`, inheriting the user's kubeconfig auth. Everything is unit-tested; the shell-out is mocked via the `Runner` interface.

**Tech Stack:** Go 1.26, [cobra](https://github.com/spf13/cobra) (scaffolded with `cobra-cli`), `gopkg.in/yaml.v3`, standard `testing`.

**Reference:** Design spec at `docs/design/2026-07-03-pier-design.md`.

---

## File Structure

```
pier/
  main.go                       # entrypoint; version vars; exit-code propagation
  cmd/
    root.go                     # root cmd, persistent flags, shared helpers (loadConfig/run/guard)
    restart.go                  # pier restart <svc>   (mutating)
    status.go                   # pier status <svc>
    logs.go                     # pier logs <svc>
    services.go                 # pier services
    config.go                   # parent "config" cmd
    config_get.go               # pier config get <svc>
    config_edit.go              # pier config edit <svc>  (mutating)
    config_path.go              # pier config path
    config_init.go              # pier config init
    secret.go                   # parent "secret" cmd
    secret_get.go               # pier secret get <svc>   (--decode)
    secret_edit.go              # pier secret edit <svc>  (mutating)
  internal/
    config/
      config.go                 # types, DefaultPath, Load, Resolve, ServiceNames
      config_test.go
    kube/
      command.go                # kubectl arg builders (pure)
      command_test.go
      runner.go                 # Runner interface + ExecRunner (real)
    confirm/
      confirm.go                # env-forward confirmation prompt
      confirm_test.go
  config.example.yaml
```

**Responsibilities:** `internal/config` owns loading + resolution (no kube knowledge). `internal/kube` owns turning a resolved `Target` into `kubectl` argument slices and running them. `internal/confirm` owns the guardrail prompt. `cmd/*` owns flag parsing and orchestration only — no business logic.

---

## Task 0: Bootstrap dependencies and cobra skeleton

**Files:**
- Modify: `go.mod`
- Create: `main.go`, `cmd/root.go` (via `cobra-cli`, overwritten in Task 5)
- Create: `config.example.yaml`

- [ ] **Step 1: Install cobra-cli and add dependencies**

Run:
```bash
go install github.com/spf13/cobra-cli@latest
go get github.com/spf13/cobra@latest
go get gopkg.in/yaml.v3@latest
```
Expected: `go.mod` gains `github.com/spf13/cobra` and `gopkg.in/yaml.v3` requires.

- [ ] **Step 2: Scaffold the cobra skeleton**

Run:
```bash
cobra-cli init
```
Expected: creates `main.go` and `cmd/root.go`. (Their contents are replaced in later tasks; this only establishes the structure and command registration wiring.)

- [ ] **Step 3: Create the example config**

Create `config.example.yaml`:
```yaml
defaultEnv: staging

environments:
  staging:
    context: staging-cluster
  prod:
    context: prod-cluster
    protected: true

services:
  api:
    namespace: default
  web:
    namespace: default
    deployment: web-server
    overrides:
      prod:
        namespace: web-prod
```

- [ ] **Step 4: Verify it builds**

Run: `go build ./...`
Expected: builds without error.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum main.go cmd/ config.example.yaml
git commit -m "chore: bootstrap cobra skeleton and dependencies"
```

---

## Task 1: Config types and resolution (`internal/config`)

**Files:**
- Create: `internal/config/config.go`
- Test: `internal/config/config_test.go`

- [ ] **Step 1: Write the failing test for Resolve**

Create `internal/config/config_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestResolve -v`
Expected: FAIL — `undefined: Config` (and related types).

- [ ] **Step 3: Write the types and Resolve**

Create `internal/config/config.go`:
```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/config/ -run TestResolve -v`
Expected: PASS (all four tests).

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config types and (service, env) resolution"
```

---

## Task 2: Config loading (`internal/config`)

**Files:**
- Modify: `internal/config/config.go`
- Test: `internal/config/load_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/config/load_test.go`:
```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("defaultEnv: staging\nenvironments:\n  staging:\n    context: staging-cluster\nservices:\n  api:\n    namespace: default\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultEnv != "staging" || cfg.Environments["staging"].Context != "staging-cluster" {
		t.Fatalf("parsed wrong: %+v", cfg)
	}
}

func TestLoadMissingFileErrors(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDefaultPathHonorsEnvVar(t *testing.T) {
	t.Setenv("PIER_CONFIG", "/custom/pier.yaml")
	if got := DefaultPath(); got != "/custom/pier.yaml" {
		t.Fatalf("got %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run "TestLoad|TestDefaultPath" -v`
Expected: FAIL — `undefined: Load` / `undefined: DefaultPath`.

- [ ] **Step 3: Add Load and DefaultPath**

Append to `internal/config/config.go`:
```go
// (add these imports to the existing import block: "os", "path/filepath",
// and "gopkg.in/yaml.v3")

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
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return &cfg, nil
}
```

Update the import block at the top of `internal/config/config.go` to:
```go
import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/config/ -v`
Expected: PASS (all tests, including Task 1's).

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: load and locate the config file"
```

---

## Task 3: kubectl argument builders (`internal/kube`)

**Files:**
- Create: `internal/kube/command.go`
- Test: `internal/kube/command_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/kube/command_test.go`:
```go
package kube

import (
	"strings"
	"testing"

	"github.com/eseceve/pier/internal/config"
)

func target() config.Target {
	return config.Target{
		Service: "api", Env: "staging",
		Context: "staging-cluster", Namespace: "default",
		Deployment: "api", ConfigMap: "api", Secret: "api",
	}
}

func joined(args []string) string { return strings.Join(args, " ") }

func TestRestartArgs(t *testing.T) {
	got := joined(RestartArgs(target()))
	want := "--context staging-cluster --namespace default rollout restart deployment/api"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLogsArgsWithFlags(t *testing.T) {
	got := joined(LogsArgs(target(), true, 100, "app"))
	want := "--context staging-cluster --namespace default logs deployment/api -f --tail=100 -c app"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLogsArgsDefaults(t *testing.T) {
	got := joined(LogsArgs(target(), false, 0, ""))
	want := "--context staging-cluster --namespace default logs deployment/api"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSecretGetArgsDecode(t *testing.T) {
	got := SecretGetArgs(target(), true)
	last := got[len(got)-1]
	if !strings.Contains(last, "base64decode") {
		t.Fatalf("decode template missing: %q", last)
	}
}

func TestConfigGetArgs(t *testing.T) {
	got := joined(ConfigGetArgs(target()))
	want := "--context staging-cluster --namespace default get configmap/api -o yaml"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/kube/ -v`
Expected: FAIL — `undefined: RestartArgs` (and others).

- [ ] **Step 3: Write the builders**

Create `internal/kube/command.go`:
```go
// Package kube builds and runs kubectl commands from a resolved Target.
package kube

import (
	"strconv"

	"github.com/eseceve/pier/internal/config"
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/kube/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/kube/
git commit -m "feat: build kubectl argument slices from a Target"
```

---

## Task 4: Runner seam (`internal/kube`)

**Files:**
- Create: `internal/kube/runner.go`

- [ ] **Step 1: Write the Runner interface and ExecRunner**

Create `internal/kube/runner.go`:
```go
package kube

import (
	"fmt"
	"os"
	"os/exec"
)

// Runner executes a kubectl invocation. It is the seam that tests replace.
type Runner interface {
	Run(args []string) error
}

// ExecRunner runs the real kubectl binary, inheriting the process stdio (so
// interactive commands like `edit` and streaming `logs -f` work).
type ExecRunner struct{}

// Run executes `kubectl <args...>`. It returns *exec.ExitError on non-zero exit
// so the caller can propagate kubectl's exit code.
func (ExecRunner) Run(args []string) error {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl not found in PATH")
	}
	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
```

- [ ] **Step 2: Verify it builds**

Run: `go build ./internal/kube/`
Expected: builds without error. (No unit test — this is a thin, side-effecting adapter; it is exercised through the mock in the cmd layer.)

- [ ] **Step 3: Commit**

```bash
git add internal/kube/runner.go
git commit -m "feat: add Runner interface and real ExecRunner"
```

---

## Task 5: Confirmation guardrail (`internal/confirm`)

**Files:**
- Create: `internal/confirm/confirm.go`
- Test: `internal/confirm/confirm_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/confirm/confirm_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/confirm/ -v`
Expected: FAIL — `undefined: Prompt`.

- [ ] **Step 3: Write the prompt**

Create `internal/confirm/confirm.go`:
```go
// Package confirm implements pier's environment-forward guardrail prompt for
// mutating operations against protected environments.
package confirm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/eseceve/pier/internal/config"
)

// Prompt shows the environment, resolved coordinates and the exact action, then
// requires the user to type the service name. It returns an error if the typed
// value does not match.
func Prompt(in io.Reader, out io.Writer, t config.Target, action string) error {
	fmt.Fprintf(out, "⚠  %s — context: %s, namespace: %s\n", strings.ToUpper(t.Env), t.Context, t.Namespace)
	fmt.Fprintf(out, "   This will: %s\n", action)
	fmt.Fprintf(out, "   Type the service name to confirm (%s): ", t.Service)

	line, _ := bufio.NewReader(in).ReadString('\n')
	if strings.TrimSpace(line) != t.Service {
		return errors.New("aborted: confirmation did not match the service name")
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/confirm/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/confirm/
git commit -m "feat: add environment-forward confirmation guardrail"
```

---

## Task 6: Root command and shared helpers (`cmd/root.go`)

**Files:**
- Modify (overwrite): `cmd/root.go`
- Modify (overwrite): `main.go`

- [ ] **Step 1: Overwrite `cmd/root.go`**

Replace the entire contents of `cmd/root.go` with:
```go
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/eseceve/pier/internal/confirm"
	"github.com/eseceve/pier/internal/config"
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
)

var (
	flagEnv     string
	flagYes     bool
	flagDryRun  bool
	flagVerbose bool

	// runner is the kubectl execution seam; overridable in tests.
	runner kube.Runner = kube.ExecRunner{}

	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:           "pier",
	Short:         "kubectl wrapper mapping service names to their k8s context/namespace",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// SetVersion injects build metadata from main.
func SetVersion(v, c, d string) {
	version, commit, date = v, c, d
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", v, c, d)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagEnv, "env", "e", "", "target environment (default: config defaultEnv)")
	rootCmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "skip confirmation for mutating ops")
	rootCmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "print the kubectl command without running it")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "print the kubectl command before running it")
}

// loadConfig loads the config from its default path.
func loadConfig() (*config.Config, error) {
	return config.Load(config.DefaultPath())
}

// resolve loads the config and resolves a service using the global --env flag.
func resolve(service string) (config.Target, error) {
	cfg, err := loadConfig()
	if err != nil {
		return config.Target{}, err
	}
	return cfg.Resolve(service, flagEnv)
}

// run honors --dry-run/--verbose, otherwise executes kubectl via the runner.
func run(args []string) error {
	if flagDryRun {
		fmt.Println("kubectl " + strings.Join(args, " "))
		return nil
	}
	if flagVerbose {
		fmt.Fprintln(os.Stderr, "+ kubectl "+strings.Join(args, " "))
	}
	return runner.Run(args)
}

// guard runs the confirmation prompt for a mutating op against a protected env,
// unless --yes was passed or --dry-run is active.
func guard(t config.Target, action string) error {
	if flagDryRun || flagYes || !t.Protected {
		return nil
	}
	return confirm.Prompt(os.Stdin, os.Stdout, t, action)
}
```

- [ ] **Step 2: Overwrite `main.go`**

Replace the entire contents of `main.go` with:
```go
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/eseceve/pier/cmd"
)

// Injected at build time via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date)
	if err := cmd.Execute(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Verify it builds**

Run: `go build ./...`
Expected: builds without error.

- [ ] **Step 4: Commit**

```bash
git add cmd/root.go main.go
git commit -m "feat: wire root command, persistent flags and shared helpers"
```

---

## Task 7: `restart` command (tracer bullet, mutating)

**Files:**
- Create (via `cobra-cli add restart`, then overwrite): `cmd/restart.go`
- Test: `cmd/restart_test.go`

- [ ] **Step 1: Generate the command file**

Run: `cobra-cli add restart`
Expected: creates `cmd/restart.go` registered on `rootCmd`.

- [ ] **Step 2: Write the failing test**

Create `cmd/restart_test.go`:
```go
package cmd

import (
	"bytes"
	"testing"
)

// fakeRunner records the args it was asked to run.
type fakeRunner struct{ got []string }

func (f *fakeRunner) Run(args []string) error { f.got = args; return nil }

// withConfig points PIER_CONFIG at a temp config and resets flags.
func withConfig(t *testing.T, yaml string) {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/config.yaml"
	if err := writeFile(path, yaml); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PIER_CONFIG", path)
	flagEnv, flagYes, flagDryRun, flagVerbose = "", false, false, false
}

const sampleYAML = `defaultEnv: staging
environments:
  staging:
    context: staging-cluster
  prod:
    context: prod-cluster
    protected: true
services:
  api:
    namespace: default
`

func TestRestartRunsRolloutRestart(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	defer func() { runner = nil }()

	rootCmd.SetArgs([]string{"restart", "api"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "--context staging-cluster --namespace default rollout restart deployment/api"
	if joinArgs(fake.got) != want {
		t.Fatalf("got %q want %q", joinArgs(fake.got), want)
	}
}
```

Create `cmd/testhelpers_test.go`:
```go
package cmd

import (
	"os"
	"strings"
)

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o600)
}

func joinArgs(args []string) string { return strings.Join(args, " ") }
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./cmd/ -run TestRestart -v`
Expected: FAIL — the generated `restart` command does not call the runner yet.

- [ ] **Step 4: Overwrite `cmd/restart.go`**

Replace the entire contents of `cmd/restart.go` with:
```go
package cmd

import (
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart <service>",
	Short: "Rollout restart of the service's deployment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		kubectlArgs := kube.RestartArgs(target)
		if err := guard(target, "kubectl "+joinArgs(kubectlArgs)); err != nil {
			return err
		}
		return run(kubectlArgs)
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
```

Note: `joinArgs` lives in `cmd/testhelpers_test.go`, which is test-only. Move it into production code — add this to `cmd/root.go` (and delete it from `testhelpers_test.go`):
```go
// joinArgs renders kubectl args for display.
func joinArgs(args []string) string { return strings.Join(args, " ") }
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./cmd/ -run TestRestart -v`
Expected: PASS.

- [ ] **Step 6: Add the guardrail test**

Append to `cmd/restart_test.go`:
```go
func TestRestartProtectedEnvBlockedWithoutConfirmation(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	defer func() { runner = nil }()
	flagEnv = "prod" // protected

	rootCmd.SetArgs([]string{"restart", "api", "-e", "prod"})
	rootCmd.SetIn(bytes.NewReader([]byte("wrong\n")))
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected confirmation to block the restart")
	}
	if fake.got != nil {
		t.Fatalf("runner should not have been called, got %v", fake.got)
	}
}
```

Note: the guardrail reads from `os.Stdin`, not cobra's input. Update `guard` in `cmd/root.go` to read from the command's input stream so it is testable — change the `confirm.Prompt` call to accept `rootCmd.InOrStdin()`:
```go
func guard(t config.Target, action string) error {
	if flagDryRun || flagYes || !t.Protected {
		return nil
	}
	return confirm.Prompt(rootCmd.InOrStdin(), os.Stdout, t, action)
}
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `go test ./cmd/ -run TestRestart -v`
Expected: PASS (both tests).

- [ ] **Step 8: Commit**

```bash
git add cmd/
git commit -m "feat: add restart command with protected-env guardrail"
```

---

## Task 8: `status` and `logs` commands

**Files:**
- Create (via `cobra-cli add`, then overwrite): `cmd/status.go`, `cmd/logs.go`
- Test: `cmd/status_logs_test.go`

- [ ] **Step 1: Generate the command files**

Run:
```bash
cobra-cli add status
cobra-cli add logs
```

- [ ] **Step 2: Write the failing tests**

Create `cmd/status_logs_test.go`:
```go
package cmd

import (
	"bytes"
	"testing"
)

// recordingRunner records every invocation (status runs two).
type recordingRunner struct{ calls [][]string }

func (r *recordingRunner) Run(args []string) error {
	r.calls = append(r.calls, args)
	return nil
}

func TestStatusRunsGetAndRolloutStatus(t *testing.T) {
	withConfig(t, sampleYAML)
	rec := &recordingRunner{}
	runner = rec
	defer func() { runner = nil }()

	rootCmd.SetArgs([]string{"status", "api"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rec.calls) != 2 {
		t.Fatalf("expected 2 kubectl calls, got %d", len(rec.calls))
	}
	if joinArgs(rec.calls[0]) != "--context staging-cluster --namespace default get deployment/api" {
		t.Fatalf("call 0 wrong: %q", joinArgs(rec.calls[0]))
	}
	if joinArgs(rec.calls[1]) != "--context staging-cluster --namespace default rollout status deployment/api" {
		t.Fatalf("call 1 wrong: %q", joinArgs(rec.calls[1]))
	}
}

func TestLogsPassesFlags(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	defer func() { runner = nil }()

	rootCmd.SetArgs([]string{"logs", "api", "-f", "--tail", "50"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "--context staging-cluster --namespace default logs deployment/api -f --tail=50"
	if joinArgs(fake.got) != want {
		t.Fatalf("got %q want %q", joinArgs(fake.got), want)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./cmd/ -run "TestStatus|TestLogs" -v`
Expected: FAIL.

- [ ] **Step 4: Overwrite `cmd/status.go`**

Replace the entire contents of `cmd/status.go` with:
```go
package cmd

import (
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <service>",
	Short: "Show the service's deployment and rollout status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		if err := run(kube.GetDeploymentArgs(target)); err != nil {
			return err
		}
		return run(kube.RolloutStatusArgs(target))
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
```

- [ ] **Step 5: Overwrite `cmd/logs.go`**

Replace the entire contents of `cmd/logs.go` with:
```go
package cmd

import (
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
)

var (
	logsFollow    bool
	logsTail      int
	logsContainer string
)

var logsCmd = &cobra.Command{
	Use:   "logs <service>",
	Short: "Show the service's deployment logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		return run(kube.LogsArgs(target, logsFollow, logsTail, logsContainer))
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "stream logs")
	logsCmd.Flags().IntVar(&logsTail, "tail", 0, "lines of recent log to show")
	logsCmd.Flags().StringVarP(&logsContainer, "container", "c", "", "container name")
	rootCmd.AddCommand(logsCmd)
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./cmd/ -run "TestStatus|TestLogs" -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add cmd/
git commit -m "feat: add status and logs commands"
```

---

## Task 9: `config` command group (`get`, `edit`, `path`, `init`)

**Files:**
- Create (via `cobra-cli add`, then overwrite): `cmd/config.go`, `cmd/config_get.go`, `cmd/config_edit.go`, `cmd/config_path.go`, `cmd/config_init.go`
- Test: `cmd/config_test.go`

- [ ] **Step 1: Generate the command files**

Run:
```bash
cobra-cli add config
cobra-cli add get -p configCmd
cobra-cli add edit -p configCmd
cobra-cli add path -p configCmd
cobra-cli add init -p configCmd
```

- [ ] **Step 2: Write the failing tests**

Create `cmd/config_test.go`:
```go
package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestConfigGetReadsConfigMap(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	defer func() { runner = nil }()

	rootCmd.SetArgs([]string{"config", "get", "api"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "--context staging-cluster --namespace default get configmap/api -o yaml"
	if joinArgs(fake.got) != want {
		t.Fatalf("got %q want %q", joinArgs(fake.got), want)
	}
}

func TestConfigInitWritesTemplate(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.yaml"
	t.Setenv("PIER_CONFIG", path)
	flagEnv, flagYes, flagDryRun, flagVerbose = "", false, false, false

	rootCmd.SetArgs([]string{"config", "init"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("config not written: %v", err)
	}
	if !strings.Contains(string(data), "defaultEnv:") {
		t.Fatalf("template missing expected content: %q", string(data))
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./cmd/ -run TestConfig -v`
Expected: FAIL.

- [ ] **Step 4: Overwrite `cmd/config.go`**

Replace the entire contents of `cmd/config.go` with:
```go
package cmd

import "github.com/spf13/cobra"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Read, edit, locate or create the config / a service's ConfigMap",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
```

- [ ] **Step 5: Overwrite `cmd/config_get.go` and `cmd/config_edit.go`**

Replace the entire contents of `cmd/config_get.go` with:
```go
package cmd

import (
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
)

var configGetCmd = &cobra.Command{
	Use:   "get <service>",
	Short: "Read the service's ConfigMap as YAML",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		return run(kube.ConfigGetArgs(target))
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
}
```

Replace the entire contents of `cmd/config_edit.go` with:
```go
package cmd

import (
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
)

var configEditCmd = &cobra.Command{
	Use:   "edit <service>",
	Short: "Edit the service's ConfigMap in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		kubectlArgs := kube.ConfigEditArgs(target)
		if err := guard(target, "kubectl "+joinArgs(kubectlArgs)); err != nil {
			return err
		}
		return run(kubectlArgs)
	},
}

func init() {
	configCmd.AddCommand(configEditCmd)
}
```

- [ ] **Step 6: Overwrite `cmd/config_path.go` and `cmd/config_init.go`**

Replace the entire contents of `cmd/config_path.go` with:
```go
package cmd

import (
	"fmt"

	"github.com/eseceve/pier/internal/config"
	"github.com/spf13/cobra"
)

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the resolved config file path",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), config.DefaultPath())
		return nil
	},
}

func init() {
	configCmd.AddCommand(configPathCmd)
}
```

Replace the entire contents of `cmd/config_init.go` with:
```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eseceve/pier/internal/config"
	"github.com/spf13/cobra"
)

const configTemplate = `defaultEnv: staging

environments:
  staging:
    context: staging-cluster
  prod:
    context: prod-cluster
    protected: true

services:
  api:
    namespace: default
  web:
    namespace: default
    deployment: web-server
    overrides:
      prod:
        namespace: web-prod
`

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Write a starter config file",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DefaultPath()
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config already exists at %s", path)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(configTemplate), 0o600); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", path)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
}
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `go test ./cmd/ -run TestConfig -v`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add cmd/
git commit -m "feat: add config get/edit/path/init commands"
```

---

## Task 10: `secret` command group (`get`, `edit`)

**Files:**
- Create (via `cobra-cli add`, then overwrite): `cmd/secret.go`, `cmd/secret_get.go`, `cmd/secret_edit.go`
- Test: `cmd/secret_test.go`

- [ ] **Step 1: Generate the command files**

Run:
```bash
cobra-cli add secret
cobra-cli add get -p secretCmd
cobra-cli add edit -p secretCmd
```
Note: cobra-cli may refuse to create a second `get.go`/`edit.go` if the file already exists from the `config` group. If so, create `cmd/secret_get.go` and `cmd/secret_edit.go` by hand with the content below.

- [ ] **Step 2: Write the failing tests**

Create `cmd/secret_test.go`:
```go
package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSecretGetPlain(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	defer func() { runner = nil }()

	rootCmd.SetArgs([]string{"secret", "get", "api"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "--context staging-cluster --namespace default get secret/api -o yaml"
	if joinArgs(fake.got) != want {
		t.Fatalf("got %q want %q", joinArgs(fake.got), want)
	}
}

func TestSecretGetDecode(t *testing.T) {
	withConfig(t, sampleYAML)
	fake := &fakeRunner{}
	runner = fake
	defer func() { runner = nil }()

	rootCmd.SetArgs([]string{"secret", "get", "api", "--decode"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(joinArgs(fake.got), "base64decode") {
		t.Fatalf("decode not applied: %q", joinArgs(fake.got))
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./cmd/ -run TestSecret -v`
Expected: FAIL.

- [ ] **Step 4: Overwrite `cmd/secret.go`**

Replace the entire contents of `cmd/secret.go` with:
```go
package cmd

import "github.com/spf13/cobra"

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Read or edit a service's Secret",
}

func init() {
	rootCmd.AddCommand(secretCmd)
}
```

- [ ] **Step 5: Overwrite `cmd/secret_get.go` and `cmd/secret_edit.go`**

Replace the entire contents of `cmd/secret_get.go` with:
```go
package cmd

import (
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
)

var secretDecode bool

var secretGetCmd = &cobra.Command{
	Use:   "get <service>",
	Short: "Read the service's Secret (use --decode for cleartext)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		return run(kube.SecretGetArgs(target, secretDecode))
	},
}

func init() {
	secretGetCmd.Flags().BoolVar(&secretDecode, "decode", false, "print secret values in cleartext")
	secretCmd.AddCommand(secretGetCmd)
}
```

Replace the entire contents of `cmd/secret_edit.go` with:
```go
package cmd

import (
	"github.com/eseceve/pier/internal/kube"
	"github.com/spf13/cobra"
)

var secretEditCmd = &cobra.Command{
	Use:   "edit <service>",
	Short: "Edit the service's Secret in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := resolve(args[0])
		if err != nil {
			return err
		}
		kubectlArgs := kube.SecretEditArgs(target)
		if err := guard(target, "kubectl "+joinArgs(kubectlArgs)); err != nil {
			return err
		}
		return run(kubectlArgs)
	},
}

func init() {
	secretCmd.AddCommand(secretEditCmd)
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./cmd/ -run TestSecret -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add cmd/
git commit -m "feat: add secret get/edit commands"
```

---

## Task 11: `services` command

**Files:**
- Create (via `cobra-cli add services`, then overwrite): `cmd/services.go`
- Test: `cmd/services_test.go`

- [ ] **Step 1: Generate the command file**

Run: `cobra-cli add services`

- [ ] **Step 2: Write the failing test**

Create `cmd/services_test.go`:
```go
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
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./cmd/ -run TestServices -v`
Expected: FAIL.

- [ ] **Step 4: Overwrite `cmd/services.go`**

Replace the entire contents of `cmd/services.go` with:
```go
package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "List configured services with their namespace and deployment",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SERVICE\tNAMESPACE\tDEPLOYMENT")
		for _, name := range cfg.ServiceNames() {
			svc := cfg.Services[name]
			deployment := svc.Deployment
			if deployment == "" {
				deployment = name
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", name, svc.Namespace, deployment)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(servicesCmd)
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./cmd/ -run TestServices -v`
Expected: PASS.

- [ ] **Step 6: Run the whole suite and lint**

Run:
```bash
go test ./... -race
golangci-lint run
```
Expected: all tests PASS; lint clean.

- [ ] **Step 7: Commit**

```bash
git add cmd/
git commit -m "feat: add services listing command"
```

---

## Task 12: Jenkinsfile (Node-free CI/CD)

**Files:**
- Create: `Jenkinsfile`

> This task encodes the pipeline designed in the spec. Before relying on it,
> verify against the real Jenkins: that an agent can run the `golang:1.26`
> container, that `GitHubJenkinsAccessToken` exists, and that GoReleaser + svu are
> reachable inside the container (install via `go install` if not baked in).

- [ ] **Step 1: Create `Jenkinsfile`**

Create `Jenkinsfile`:
```groovy
pipeline {
  agent any

  options {
    timeout(time: 20, unit: 'MINUTES')
  }

  environment {
    APP_NAME = "pier"
    GO_IMAGE = "golang:1.26"
    GITHUB_TOKEN = credentials("GitHubJenkinsAccessToken")
  }

  stages {
    stage("Test") {
      when { not { anyOf { branch "main"; branch "develop" } } }
      steps {
        script {
          go_run("go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest && golangci-lint run")
          go_run("go test ./... -race")
          go_run("go build ./...")
        }
      }
    }

    stage("Release") {
      when { anyOf { branch "main"; branch "develop" } }
      steps {
        script {
          release()
        }
      }
    }
  }

  post {
    always {
      script {
        jenkinsNotification()
      }
    }
  }
}

def go_run(cmd) {
  sh "docker run --rm -v ${WORKSPACE}:/app -w /app ${GO_IMAGE} sh -c '${cmd}'"
}

def release() {
  sh 'git config --global user.email "jenkins@comparaonline.com"'
  sh 'git config --global user.name "Jenkins"'

  def prerelease = env.BRANCH_NAME == 'develop' ? '--pre-release beta' : ''
  def next = sh(script: "docker run --rm -v ${WORKSPACE}:/app -w /app ${GO_IMAGE} sh -c 'go install github.com/caarlos0/svu/v3@latest && svu next ${prerelease}'", returnStdout: true).trim()
  def current = sh(script: "docker run --rm -v ${WORKSPACE}:/app -w /app ${GO_IMAGE} sh -c 'go install github.com/caarlos0/svu/v3@latest && svu current'", returnStdout: true).trim()

  if (next == current) {
    echo "No release warranted (svu next == current: ${current})"
    return
  }

  sh "git tag ${next}"
  sh "git push origin ${next}"
  go_run("go install github.com/goreleaser/goreleaser/v2@latest && GITHUB_TOKEN=${GITHUB_TOKEN} goreleaser release --clean")

  if (env.BRANCH_NAME == 'main') {
    sh "git checkout develop || git checkout -b develop origin/develop"
    sh "git merge --strategy-option=ours origin/main"
    sh "git push origin develop"
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add Jenkinsfile
git commit -m "ci: add Node-free Jenkins pipeline (svu + GoReleaser)"
```

---

## Task 13: README quickstart verification

**Files:**
- Modify: `README.md` (only if commands drift from the implementation)

- [ ] **Step 1: Smoke-test the CLI end to end with --dry-run**

Run:
```bash
go run . config init
go run . restart api --dry-run
go run . logs api -f --dry-run
go run . config get api --dry-run
go run . services
```
Expected: `config init` writes the config; `--dry-run` prints the exact `kubectl ...` lines; `services` prints the table.

- [ ] **Step 2: Reconcile README**

Confirm the commands and flags in `README.md` match the implemented behavior. Fix any drift.

- [ ] **Step 3: Commit (if changed)**

```bash
git add README.md
git commit -m "docs: reconcile README with implemented CLI"
```

---

## Self-Review Notes

- **Spec coverage:** commands (restart/status/logs/config get·edit·path·init/secret get·edit/services) → Tasks 7–11; resolution + defaults + per-env override → Task 1; config location/`PIER_CONFIG` → Task 2; kubectl builders incl. `--decode` → Task 3; Runner + exit-code propagation → Tasks 4 & 6; env-forward guardrail + `--yes`/`--dry-run` → Tasks 5–7; global flags → Task 6; CI/CD (Jenkins + svu + GoReleaser) → Task 12; distribution/`go install` via ldflags version → Tasks 6 & 12.
- **Deferred by design:** Homebrew tap (later version); committed `CHANGELOG.md` (GitHub Release notes are the changelog).
- **Type consistency:** `config.Target` fields and `kube.*Args(config.Target)` signatures are shared across Tasks 1, 3, 7–11; `runner` seam and `joinArgs`/`resolve`/`run`/`guard` helpers are defined once in Task 6.
