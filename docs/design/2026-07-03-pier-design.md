# pier — Design

**Status:** Approved (design phase)
**Date:** 2026-07-03

## Summary

`pier` is a small command-line tool that wraps `kubectl`. It maps a friendly
service name to its Kubernetes coordinates (context + namespace, plus optional
resource names) via a config file, so day-to-day operations become short and
memorable. For example:

```
pier restart ai-interactor
```

resolves `ai-interactor` to its `(context, namespace, deployment)` and runs the
equivalent `kubectl rollout restart` against **staging by default**.

The name comes from the nautical theme of Kubernetes (Greek *kubernetes* =
"helmsman"; the logo is a ship's wheel): a *pier* is where ships dock, mirroring
how each service is "docked" at a known context/namespace.

## Goals

- Turn common multi-flag `kubectl` invocations into short, service-oriented
  commands.
- Default to a safe environment (staging); reach other environments explicitly.
- Guard mutating operations against protected environments (e.g. production).
- Be a single static binary, easy to distribute (OSS, own Homebrew tap later).

## Non-goals (v1)

- No `exec`/interactive shell into pods.
- No replacement for `kubectl`; anything not modeled falls back to the user
  running `kubectl` directly.
- No cluster provisioning, no manifest management, no CI integration.

## Architecture

- **Language / framework:** Go + [cobra](https://github.com/spf13/cobra) (the
  same CLI framework `kubectl` uses). Single static binary named `pier`.
- **Execution model:** shell-out to `kubectl`. `pier` builds the `kubectl`
  command with resolved `--context` and `--namespace` flags and executes it,
  inheriting stdio. This inherits the user's existing kubeconfig auth — important
  because EKS contexts authenticate via an exec plugin (`aws eks get-token`),
  which would be costly to replicate with `client-go`.
- **Startup check:** verify `kubectl` is present in `PATH`; if not, exit with a
  clear error.

### Rejected alternative

- **`client-go` (native Kubernetes SDK):** full programmatic control, but must
  replicate EKS exec-based auth and is far more code than a wrapper needs.
  Rejected for v1.

## Configuration

- **Format:** YAML.
- **Location:** `$XDG_CONFIG_HOME/pier/config.yaml` (fallback
  `~/.config/pier/config.yaml`). Override with the `PIER_CONFIG` environment
  variable.
- **DRY schema:** the context is declared once per environment; services only
  reference a namespace and, optionally, resource names that differ from the
  service name.

```yaml
defaultEnv: staging

environments:
  staging:
    context: secontreras@k8s-staging.us-east-1.eksctl.io
  prod:
    context: scontreras@co-applications.us-east-1.eksctl.io
    protected: true            # triggers confirmation on mutating ops

services:
  ai-interactor:
    namespace: default
    # deployment / configmap / secret are optional; default = service name
  conversations:
    namespace: default
    deployment: conversations-api      # override when the deployment differs
    overrides:                         # only when something changes per env
      prod:
        namespace: conversations-prod
```

### Resolution rules

Given `(service, env)`:

1. `env` defaults to `defaultEnv` when `-e/--env` is not passed.
2. `context` comes from `environments[env].context`.
3. `namespace` = `services[svc].overrides[env].namespace` if present, else
   `services[svc].namespace`.
4. `deployment` / `configmap` / `secret` = the service's field if set, else the
   service name.
5. Unknown service or missing mapping for the requested env → clear error.

## Commands (v1)

`<svc>` is a service key from the config. Staging is the default target.

| Command | Runs (conceptually) |
|---|---|
| `pier restart <svc>` | `kubectl rollout restart deployment/<dep>` |
| `pier status <svc>` | `kubectl get pods` + `kubectl rollout status deployment/<dep>` |
| `pier logs <svc>` | `kubectl logs deployment/<dep>` (flags `-f/--follow`, `--tail`, `-c/--container`) |
| `pier config get <svc>` | `kubectl get configmap/<cm> -o yaml` |
| `pier config edit <svc>` | `kubectl edit configmap/<cm>` (opens `$EDITOR`) |
| `pier secret get <svc>` | `kubectl get secret/<sec> -o yaml` (flag `--decode` for base64) |
| `pier secret edit <svc>` | `kubectl edit secret/<sec>` |
| `pier services` | List configured services with their namespace/deployment |
| `pier config init` | Write a starter config template |
| `pier config path` | Print the resolved config file path |

Every resolved command is invoked with `--context <ctx> --namespace <ns>`.

### Global flags

- `-e, --env <env>` — target environment (default: `staging`).
- `-y, --yes` — skip the confirmation prompt for mutating ops.
- `--dry-run` — print the `kubectl` command that would run, without executing.
- `-v, --verbose` — print the `kubectl` command before executing it.

## Environments and guardrails

- Without `-e`, every command targets `staging`.
- **Mutating** operations (`restart`, `config edit`, `secret edit`) against an
  environment with `protected: true` require an interactive confirmation
  (typing the service name). `-y/--yes` skips it.
- **Read** operations (`logs`, `status`, `*get`, `services`) never prompt.

### Security note

`secret get --decode` prints secret values in cleartext (they remain in shell
scrollback). It is opt-in via the `--decode` flag; the default prints the raw
(base64) resource as `kubectl` does.

## Error handling

- Unknown service → error listing near matches / available services.
- Unknown environment, or service with no mapping for the requested env → clear
  error naming what is missing.
- `kubectl` missing from `PATH` → error at startup.
- `pier` propagates `kubectl`'s exit code unchanged, so it composes in scripts.

## Testing

- **Unit tests** for the pure logic: config parsing, `(svc, env) → target`
  resolution, and command construction. `--dry-run` makes the built command
  assertable as a string.
- The shell-out is isolated behind a `Runner` interface; tests inject a mock,
  and `main` wires the real exec-based runner.

## Project structure

```
pier/
  main.go
  cmd/            # cobra commands: root, restart, logs, status, config, secret, services
  internal/
    config/       # load + schema + (svc, env) -> target resolution
    kube/         # command builder + Runner (real exec / mock)
  config.example.yaml
  README.md
  LICENSE
```

## Distribution (later)

- Public GitHub repo (own namespace, decoupled from any employer).
- Homebrew tap + `goreleaser` for static binaries. Out of scope for the first
  implementation plan.
