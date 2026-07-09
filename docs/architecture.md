# Architecture

`pier` is a **thin `kubectl` wrapper**, not a Kubernetes client. It resolves a
friendly service name to a context/namespace (plus resource names) from config,
builds a `kubectl` argument list, and execs `kubectl` — inheriting the user's
stdio and propagating its exit code. The rationale is in the
[design spec](design/2026-07-03-pier-design.md); this page summarizes the shape.

## Guardrails

These decisions keep `pier` small; they are load-bearing, not incidental.

- **Shell out to `kubectl`; don't use `client-go`.** Exec-ing `kubectl` inherits
  the user's existing kubeconfig authentication for free — which matters because
  EKS contexts authenticate through an exec plugin (`aws eks get-token`).
  Adopting the Kubernetes SDK would mean replicating that auth and far more code
  than a wrapper needs.
- **Don't parse or normalize `kubectl` output.** `kubectl`'s stdout/stderr
  stream through untouched and its exit code is propagated. `pier` adds
  resolution and ergonomics, not a new output format — parsing `kubectl`'s text
  would be brittle and couple `pier` to its formatting.
- **Fall back to `kubectl`, don't reimplement it.** Anything not modeled by a
  `pier` command is out of scope; the user runs `kubectl` directly. In
  particular, no `exec`/interactive shell into pods.
- **Verify `kubectl` is on `PATH`.** The runner checks for `kubectl` before
  executing and fails with a clear error if it is missing.
- **Confirm before a protected blast radius.** Mutating operations against a
  `protected` environment prompt with the resolved context/namespace and the
  exact `kubectl` command, so the impact is visible before it runs.

## Command flow

Every operation follows the same path:

1. **Resolve** — load the config and turn `(service, --env)` into a `Target`
   (context, namespace, deployment/configmap/secret, protected flag). See
   [Configuration](configuration.md) for precedence.
2. **Build args** — assemble the `kubectl` argument slice for the operation,
   with `--context` and `--namespace` from the `Target`.
3. **Guard** — for mutating ops on a protected env, prompt for confirmation
   (skipped by `-y/--yes` or `--dry-run`).
4. **Run** — honor `--dry-run` (print only) and `--verbose` (print then run),
   otherwise exec `kubectl`, inheriting stdio and propagating the exit code.

## Package layout

| Path | Responsibility |
|---|---|
| `main.go` | Entry point; injects build metadata and calls `cmd.Execute`. |
| `cmd/` | Cobra commands, global flags, and the resolve/guard/run helpers. |
| `internal/config/` | Config types, path resolution, and `(service, env)` → `Target`. |
| `internal/kube/` | `kubectl` argument builders (`command.go`) and the exec seam (`runner.go`). |
| `internal/confirm/` | The protected-environment confirmation prompt. |

The `kube.Runner` interface is the execution seam: the real `ExecRunner` shells
out to `kubectl`, and tests substitute a fake runner to assert on the arguments
`pier` builds without touching a cluster.

## Documentation generation

The per-command reference under [`commands/`](commands/README.md) is generated
from the cobra command tree (`make docs`), never hand-written, so it can't drift
from the code. Man pages and shell completions derive from the same tree.

## Decision records

Architecture decisions are captured as ADRs under `docs/adr/` as they are made
(the first one lands with the agent-consumable work). Together with the
[design spec](design/2026-07-03-pier-design.md) and
[implementation plans](plans/), they record why `pier` is shaped this way.
