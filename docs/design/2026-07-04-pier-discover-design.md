# pier `discover` — Design

**Status:** Approved (design phase)
**Date:** 2026-07-04
**Base branch:** `feat/cli-implementation` (this feature depends on the pier CLI code not yet on `main`).

## Summary

Auto-generate pier's `config.yaml` from a live cluster instead of hand-writing
it. Two interactive front-ends over one shared discovery core, all generic (no
company-specific values baked into the OSS binary):

1. **`internal/discover`** — a Go core that queries `kubectl` and returns
   structured data + resource-name matching.
2. **`pier discover`** — a CLI command with an interactive wizard (default) and a
   non-interactive `--json` data mode.
3. **A repo-shipped Claude Code skill** — drives the same flow conversationally
   for Claude users, consuming `pier discover --json`.

Company-specific conventions (e.g. secrets named `<service>-local`) are never
hardcoded — they are *discovered* by scanning the cluster and matching names.

## Goals

- Turn "hand-write YAML mapping services to context/namespace/secret" into a
  guided flow driven by what actually exists in the cluster.
- Keep the discovery logic in one place, reused by both front-ends.
- Stay generic: usable by any dev on any cluster, inside or outside Comparaonline.
- Never clobber an existing config without confirmation/backup.

## Non-goals

- Inferring which context is "prod" vs "staging" — that is a human decision the
  flow asks for.
- Discovering resources pier does not model (jobs, cronjobs, ingresses).
- A general kubectl explorer — scope is building pier config only.

## Architecture

### 1. `internal/discover` (shared core)

Pure-ish functions over a capturing command runner (so tests mock kubectl):

```go
type CmdRunner interface { Output(args []string) ([]byte, error) }

func Contexts(r CmdRunner) ([]string, error)              // kubectl config get-contexts -o name
func Namespaces(r CmdRunner, ctx string) ([]string, error)
func Deployments(r CmdRunner, ctx, ns string) ([]string, error)
func ConfigMaps(r CmdRunner, ctx, ns string) ([]string, error)
func Secrets(r CmdRunner, ctx, ns string) ([]string, error)

// Match picks the best candidate name for a service: exact match, else a single
// "<service>-*" / "<service>*" prefix match, else "" (no confident match).
func Match(service string, candidates []string) string
```

- `CmdRunner` is separate from `kube.Runner` because discovery needs **captured
  stdout**, whereas `kube.Runner` inherits stdio for interactive/streaming ops.
  The real impl uses `exec.Command(...).Output()`.
- `Match` filters out noise (helm release secrets `sh.helm.release.*`, service
  account tokens) before matching.

### 2. `pier discover` command

**Interactive wizard (default):**

1. `Contexts()` → prompt the user to select which contexts to register, and for
   each: an environment name (default: a slug of the context) and `protected`
   (y/N).
2. Per environment: list namespaces, let the user pick one (or type it); list
   deployments in it; let the user select which to include as services.
3. Per service: default `deployment` to the name; run `Match` against ConfigMaps
   and Secrets and show the guess — the user accepts, overrides, or skips.
4. Correlate services that appear under the same name across environments; when a
   field (namespace/secret/configmap) differs per env, emit it under `overrides`.
5. Render `config.yaml`. If one exists at `config.DefaultPath()`, show what would
   change and back it up (`config.yaml.bak`) before writing, on confirmation.

**`--json` (non-interactive data mode):** emit discovered data as JSON for
scripting and for the skill. Shape:
- `pier discover --json` → `{ "contexts": [...] }`
- `pier discover --json --context X --namespace ns` →
  `{ "deployments": [...], "configmaps": [...], "secrets": [...] }`

The pure assembly step (services + overrides → config YAML) lives in a testable
function shared by the wizard.

### 3. Repo-shipped skill

A generic Claude Code skill under `skills/` in the repo (so `git clone` +
install makes it available). It:
- calls `pier discover --json` for raw data,
- converses with the user to choose contexts/env names/services,
- applies LLM judgment for fuzzy secret/configmap matching and cross-env
  correlation,
- writes `~/.config/pier/config.yaml` (respecting backup-on-exists).

It reimplements **no** kubectl parsing — all data comes from `pier discover`.

## Build order

Phase 1 (this plan's focus): `internal/discover` core + `pier discover --json`
data mode + the skill. Delivers the full conversational value.

Phase 2 (follow-up): the interactive Go wizard (`pier discover` with prompts),
for CLI-only users without Claude.

## Error handling

- `kubectl` missing → clear error (reuse existing check).
- Unreachable context / RBAC denied → surface kubectl's stderr, continue the flow
  for other contexts where possible.
- Existing config → never overwrite silently; back up + confirm.

## Testing

- `internal/discover`: unit-test `Match` (exact/prefix/none, noise filtering) and
  each list function's output parsing via a mock `CmdRunner`.
- `pier discover --json`: assert the emitted JSON shape with a mock runner.
- The YAML assembly function: table tests (single env, multi-env with overrides,
  matched vs skipped secret).
- Interactive wizard prompts (phase 2) are exercised through the pure assembly
  function; raw prompt I/O is kept thin.
