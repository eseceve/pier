# pier `discover` (Phase 1) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add cluster discovery to pier so the config can be generated from a live cluster — Phase 1: a shared Go discovery core, a `pier discover --json` data command, and a repo-shipped Claude Code skill that drives the flow conversationally.

**Architecture:** `internal/discover` holds pure helpers (`parseNames`, `filterNoise`, `Match`) plus a `CmdRunner` seam and kubectl list functions. `pier discover --json` exposes the data as a drill-down (contexts → namespaces → resources). A `.claude/skills/pier-config` skill consumes that JSON and writes `~/.config/pier/config.yaml`. The interactive Go wizard is deferred to Phase 2.

**Tech Stack:** Go 1.26, cobra, standard `testing`, `encoding/json`.

**Reference:** `docs/design/2026-07-04-pier-discover-design.md`. Base branch: `feat/config-discover`.

---

## File Structure

```
internal/discover/
  discover.go        # CmdRunner + ExecRunner + list funcs (Contexts/Namespaces/Deployments/ConfigMaps/Secrets)
  match.go           # pure helpers: parseNames, filterNoise, Match
  match_test.go      # tests for the pure helpers
  discover_test.go   # tests for list funcs via a mock CmdRunner
cmd/
  discover.go        # `pier discover` (--json data mode; non-json prints guidance)
  discover_test.go   # tests via a mock discover.CmdRunner
.claude/skills/pier-config/
  SKILL.md           # generic skill that builds the config from `pier discover --json`
README.md            # document `pier discover` + the skill
```

---

## Task 1: Pure discovery helpers (`internal/discover/match.go`)

**Files:**
- Create: `internal/discover/match.go`
- Test: `internal/discover/match_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/discover/match_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/discover/ -v`
Expected: FAIL — `undefined: parseNames` / `filterNoise` / `Match`.

- [ ] **Step 3: Write the implementation**

Create `internal/discover/match.go`:
```go
// Package discover inspects a Kubernetes cluster via kubectl to help build
// pier's config.
package discover

import "strings"

// parseNames splits `kubectl ... -o name` output into bare names, stripping any
// "kind/" prefix and blank lines.
func parseNames(out []byte) []string {
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if i := strings.LastIndex(line, "/"); i >= 0 {
			line = line[i+1:]
		}
		names = append(names, line)
	}
	return names
}

// filterNoise drops Kubernetes-managed names that never map to a pier service:
// helm release secrets, service-account tokens, and cluster CA material.
func filterNoise(names []string) []string {
	var out []string
	for _, n := range names {
		switch {
		case strings.HasPrefix(n, "sh.helm.release."):
		case strings.Contains(n, "-token-"):
		case n == "kube-root-ca.crt":
		case strings.HasPrefix(n, "istio-ca"):
		default:
			out = append(out, n)
		}
	}
	return out
}

// Match returns the best candidate name for a service: an exact match, else the
// unique candidate prefixed with "<service>-", else "".
func Match(service string, candidates []string) string {
	for _, c := range candidates {
		if c == service {
			return c
		}
	}
	var prefixed []string
	for _, c := range candidates {
		if strings.HasPrefix(c, service+"-") {
			prefixed = append(prefixed, c)
		}
	}
	if len(prefixed) == 1 {
		return prefixed[0]
	}
	return ""
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/discover/ -v`
Expected: PASS (all seven).

- [ ] **Step 5: Lint**

Run: `golangci-lint run ./internal/discover/...` (use `$(go env GOPATH)/bin/golangci-lint` if not on PATH). Expected: 0 issues.

- [ ] **Step 6: Commit**

```bash
git add internal/discover/
git -c user.name="Sebastian Contreras" -c commit.gpgsign=false commit -m "feat: add discovery name parsing, noise filtering and matching"
```

---

## Task 2: kubectl list functions (`internal/discover/discover.go`)

**Files:**
- Create: `internal/discover/discover.go`
- Test: `internal/discover/discover_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/discover/discover_test.go`:
```go
package discover

import (
	"reflect"
	"strings"
	"testing"
)

// fakeRunner returns canned output keyed by the joined kubectl args.
type fakeRunner struct{ out map[string]string }

func (f fakeRunner) Output(args []string) ([]byte, error) {
	return []byte(f.out[strings.Join(args, " ")]), nil
}

func TestContexts(t *testing.T) {
	r := fakeRunner{out: map[string]string{
		"config get-contexts -o name": "staging-ctx\nprod-ctx\n",
	}}
	got, err := Contexts(r)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"staging-ctx", "prod-ctx"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestDeployments(t *testing.T) {
	r := fakeRunner{out: map[string]string{
		"--context c --namespace n get deployments -o name": "deployment.apps/api\ndeployment.apps/web\n",
	}}
	got, err := Deployments(r, "c", "n")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"api", "web"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestSecretsAreFiltered(t *testing.T) {
	r := fakeRunner{out: map[string]string{
		"--context c --namespace n get secrets -o name": "secret/api-local\nsecret/sh.helm.release.v1.api.v1\n",
	}}
	got, err := Secrets(r, "c", "n")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"api-local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/discover/ -run "TestContexts|TestDeployments|TestSecrets" -v`
Expected: FAIL — `undefined: Contexts` / `Deployments` / `Secrets` / `Output`.

- [ ] **Step 3: Write the implementation**

Create `internal/discover/discover.go`:
```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/discover/ -v`
Expected: PASS (all tests, including Task 1's).

- [ ] **Step 5: Lint**

Run: `golangci-lint run ./internal/discover/...`
Expected: 0 issues. (If gosec still flags G204 despite the annotation, keep the annotation; it is justified.)

- [ ] **Step 6: Commit**

```bash
git add internal/discover/
git -c user.name="Sebastian Contreras" -c commit.gpgsign=false commit -m "feat: add kubectl-backed discovery list functions"
```

---

## Task 3: `pier discover` command (`cmd/discover.go`)

**Files:**
- Create: `cmd/discover.go`
- Test: `cmd/discover_test.go`

- [ ] **Step 1: Write the failing test**

Create `cmd/discover_test.go`:
```go
package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/eseceve/pier/internal/discover"
)

// fakeCmdRunner returns canned kubectl output keyed by joined args.
type fakeCmdRunner struct{ out map[string]string }

func (f fakeCmdRunner) Output(args []string) ([]byte, error) {
	return []byte(f.out[strings.Join(args, " ")]), nil
}

func withDiscover(t *testing.T, r discover.CmdRunner) {
	t.Helper()
	prev := cmdRunner
	cmdRunner = r
	discoverJSON, discoverContext, discoverNamespace = false, "", ""
	t.Cleanup(func() { cmdRunner = prev })
}

func TestDiscoverJSONListsContexts(t *testing.T) {
	withDiscover(t, fakeCmdRunner{out: map[string]string{
		"config get-contexts -o name": "staging-ctx\nprod-ctx\n",
	}})

	rootCmd.SetArgs([]string{"discover", "--json"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, `"contexts"`) || !strings.Contains(s, "staging-ctx") {
		t.Fatalf("got %q", s)
	}
}

func TestDiscoverJSONListsResourcesFiltered(t *testing.T) {
	withDiscover(t, fakeCmdRunner{out: map[string]string{
		"--context c --namespace n get deployments -o name": "deployment.apps/api\n",
		"--context c --namespace n get configmaps -o name":  "configmap/api-config\n",
		"--context c --namespace n get secrets -o name":     "secret/api-local\nsecret/sh.helm.release.v1.api.v1\n",
	}})

	rootCmd.SetArgs([]string{"discover", "--json", "--context", "c", "--namespace", "n"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "api-local") || !strings.Contains(s, "api-config") {
		t.Fatalf("missing resources: %q", s)
	}
	if strings.Contains(s, "helm") {
		t.Fatalf("helm secret should be filtered: %q", s)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -run TestDiscover -v`
Expected: FAIL — no `discover` command / undefined `cmdRunner`.

- [ ] **Step 3: Write the implementation**

Create `cmd/discover.go`:
```go
package cmd

import (
	"encoding/json"

	"github.com/eseceve/pier/internal/discover"
	"github.com/spf13/cobra"
)

var (
	discoverJSON      bool
	discoverContext   string
	discoverNamespace string

	// cmdRunner is the discovery seam; overridable in tests.
	cmdRunner discover.CmdRunner = discover.ExecRunner{}
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover cluster resources to help build the config",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if !discoverJSON {
			_, err := cmd.OutOrStdout().Write([]byte(
				"Interactive discovery is coming in a later version.\n" +
					"For now, use `pier discover --json [--context X --namespace N]`,\n" +
					"or the pier-config skill to build your config conversationally.\n"))
			return err
		}

		out := map[string]any{}
		switch {
		case discoverContext == "":
			ctxs, err := discover.Contexts(cmdRunner)
			if err != nil {
				return err
			}
			out["contexts"] = ctxs
		case discoverNamespace == "":
			nss, err := discover.Namespaces(cmdRunner, discoverContext)
			if err != nil {
				return err
			}
			out["namespaces"] = nss
		default:
			deps, err := discover.Deployments(cmdRunner, discoverContext, discoverNamespace)
			if err != nil {
				return err
			}
			cms, err := discover.ConfigMaps(cmdRunner, discoverContext, discoverNamespace)
			if err != nil {
				return err
			}
			secs, err := discover.Secrets(cmdRunner, discoverContext, discoverNamespace)
			if err != nil {
				return err
			}
			out["deployments"] = deps
			out["configmaps"] = cms
			out["secrets"] = secs
		}

		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	},
}

func init() {
	discoverCmd.Flags().BoolVar(&discoverJSON, "json", false, "emit discovered data as JSON")
	discoverCmd.Flags().StringVar(&discoverContext, "context", "", "context to inspect")
	discoverCmd.Flags().StringVar(&discoverNamespace, "namespace", "", "namespace to inspect (requires --context)")
	rootCmd.AddCommand(discoverCmd)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd/ -run TestDiscover -v`
Expected: PASS (both tests).

- [ ] **Step 5: Full check + lint**

Run:
```bash
go test ./... -race
golangci-lint run ./...
```
Expected: all pass; 0 lint issues.

- [ ] **Step 6: Commit**

```bash
git add cmd/
git -c user.name="Sebastian Contreras" -c commit.gpgsign=false commit -m "feat: add discover command with --json data mode"
```

---

## Task 4: Repo-shipped skill (`.claude/skills/pier-config/SKILL.md`)

**Files:**
- Create: `.claude/skills/pier-config/SKILL.md`

- [ ] **Step 1: Create the skill**

Create `.claude/skills/pier-config/SKILL.md`:
```markdown
---
name: pier-config
description: Build or update a pier CLI config.yaml by discovering Kubernetes resources with the `pier discover` command. Use when the user wants to generate, bootstrap, or update their pier config from a live cluster — mapping services to their context, namespace, deployment, ConfigMap and Secret. Triggers on "build my pier config", "generate the pier config", "bootstrap pier from the cluster", "armá el config de pier", "generá el pier config".
---

# Building a pier config from a cluster

Generate `~/.config/pier/config.yaml` (or the path from `pier config path`) by
discovering what actually exists in the user's cluster. Use `pier discover --json`
for all cluster data — never parse kubectl yourself. This skill is generic: infer
naming conventions from what you find, do not assume any specific one.

## Steps

1. **Locate the config.** Run `pier config path`. If a file exists there, read it
   and treat this as an update; otherwise you are creating it. Never overwrite
   without showing the diff and confirming; back up an existing file to
   `config.yaml.bak` first.

2. **Environments.** Run `pier discover --json` to list contexts. Ask the user
   which contexts to register and, for each, an environment name (e.g. `staging`,
   `prod`) and whether it is `protected: true` (production-like). Pick a sensible
   `defaultEnv`.

3. **Services per environment.** For each environment's context, run
   `pier discover --json --context <ctx>` to list namespaces and ask which to
   scan (or ask the user directly). Then
   `pier discover --json --context <ctx> --namespace <ns>` to get deployments,
   configmaps, and secrets. Ask which deployments to include as services.

4. **Match resources.** For each chosen service, default `deployment` to the
   service name. Look at the returned `configmaps`/`secrets` and pick the best
   match by name (exact, else a single `<service>-*` candidate). Show your guess
   and let the user confirm, override, or skip. Only emit `configmap:`/`secret:`
   when the name differs from the service name.

5. **Correlate across environments.** When the same service appears in multiple
   environments and a field (namespace/secret/configmap) differs, put the
   non-default value under `overrides.<env>`.

6. **Write and verify.** Write the YAML, then run `pier services` and a couple of
   `pier <cmd> <svc> --dry-run` invocations to confirm the config resolves as
   expected. Report what was written.

## Notes

- If a context is unreachable (VPN/RBAC), say so and continue with the others.
- Services with no matching ConfigMap/Secret are normal — leave those fields off;
  `config get`/`secret get` simply will not apply to them.
```

- [ ] **Step 2: Verify the skill is well-formed**

Run: `head -5 .claude/skills/pier-config/SKILL.md`
Expected: valid YAML frontmatter with `name:` and `description:`.

- [ ] **Step 3: Commit**

```bash
git add .claude/skills/pier-config/SKILL.md
git -c user.name="Sebastian Contreras" -c commit.gpgsign=false commit -m "feat: add pier-config skill to build config from cluster"
```

---

## Task 5: Document and finalize

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add a discovery section to `README.md`**

Insert after the `## Configuration` section in `README.md`:
```markdown
## Generating the config

Build the config from a live cluster instead of writing it by hand.

Raw discovery data (drill down contexts → namespaces → resources):

```console
pier discover --json                                  # list contexts
pier discover --json --context <ctx>                  # list namespaces
pier discover --json --context <ctx> --namespace <ns> # deployments, configmaps, secrets
```

Or, in Claude Code, use the **pier-config** skill (shipped in
`.claude/skills/pier-config/`) to build the config conversationally — it drives
`pier discover` and writes the config for you.

> An interactive `pier discover` wizard (no Claude required) is planned for a
> later version.
```

- [ ] **Step 2: Smoke-check the command builds and the data mode runs**

Run:
```bash
go build ./...
go run . discover --json | head -c 200
```
Expected: builds; prints a JSON object with a `contexts` array (real, from the user's kubeconfig).

- [ ] **Step 3: Full verification**

Run:
```bash
go test ./... -race
golangci-lint run ./...
```
Expected: all pass; 0 lint issues.

- [ ] **Step 4: Commit**

```bash
git add README.md
git -c user.name="Sebastian Contreras" -c commit.gpgsign=false commit -m "docs: document pier discover and the pier-config skill"
```

---

## Self-Review Notes

- **Spec coverage:** discovery core (`Contexts/Namespaces/Deployments/ConfigMaps/Secrets`, `Match`, noise filtering, `CmdRunner` seam) → Tasks 1–2; `pier discover --json` drill-down data mode → Task 3; repo-shipped generic skill consuming `pier discover --json` → Task 4; docs → Task 5. Non-goals respected (no prod inference, no unmodeled resources).
- **Deferred by design (Phase 2):** interactive Go wizard — Task 3 prints guidance when `--json` is absent, leaving room for the wizard to become the default later.
- **Type consistency:** `discover.CmdRunner` / `Output([]string) ([]byte, error)` is used identically in Tasks 2–3; the `cmdRunner` seam mirrors the existing `runner` (kube) pattern; `Match` signature is shared across tasks.
