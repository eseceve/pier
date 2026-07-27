# pier — agent guide

`pier` is a small `kubectl` wrapper CLI (Go module `github.com/eseceve/pier`).
It maps a friendly service name to its Kubernetes coordinates (context +
namespace, plus optional resource names) via a config file, so common operations
become short and memorable — **targeting `staging` by default**.

This file is the source of truth for how to work in this repo; `CLAUDE.md` is a
symlink to it. Read `README.md` for the user-facing surface, `CONTRIBUTING.md` for
the full dev/release workflow, and `docs/design/` for the design rationale.

## Architecture guardrails

These are the decisions that keep `pier` small. Preserve them; don't "improve" the
tool into something larger.

- **Thin wrapper, not a client.** `pier` shells out to `kubectl`: it resolves
  `--context`/`--namespace` from config and execs `kubectl`, inheriting stdio.
  Do **not** reach for `client-go` / the Kubernetes SDK — shelling out inherits the
  user's existing kubeconfig auth for free, which matters because EKS contexts
  authenticate via an exec plugin (`aws eks get-token`). Adding `client-go` would
  mean replicating that auth and far more code than a wrapper needs.
- **Don't normalize or parse `kubectl` output.** Stream `kubectl`'s stdout/stderr
  through untouched and **propagate its exit code**. `pier` adds resolution and
  ergonomics, not a new output format — parsing `kubectl`'s text would be brittle
  and couple us to its formatting.
- **Fall back to `kubectl`, don't reimplement it.** Anything not modeled by a
  `pier` command is out of scope — the user runs `kubectl` directly. In particular:
  **no** `exec`/interactive shell into pods, and no attempt to be a `kubectl`
  replacement.
- **Startup check:** verify `kubectl` is on `PATH`; if missing, exit with a clear
  error (don't try to continue).
- **Safety on mutating ops.** `staging` is the default target. Mutating operations
  (`restart`, `config edit`, `secret edit`) against a `protected` environment must
  prompt for confirmation showing the resolved context/namespace and the exact
  `kubectl` action, so the blast radius is visible. `-y/--yes` skips the prompt;
  `--dry-run` prints the `kubectl` command without running it; `-v/--verbose` prints
  it before running.

## Config model

- **Location:** `$XDG_CONFIG_HOME/pier/config.yaml` (fallback
  `~/.config/pier/config.yaml`); override with `PIER_CONFIG`.
- **Resolution:** a service resolves to `deployment`/`configmap`/`secret` = its
  configured field if set, else the service name. Per-environment `overrides` can
  change **namespace, configmap, and secret** (not just namespace) so the same
  service can differ across environments (e.g. `web` → namespace `web-prod` on
  prod). Precedence is env override → service base → service-name default.

## Development

Requires **Go 1.26+** and `kubectl` on `PATH`. Common commands (see `Makefile`):

```console
make build   # ./bin/pier
make test    # go test ./... -race -cover
make lint    # golangci-lint run
make fmt     # gofmt -w .
```

Install git hooks once with `lefthook install` — they run `commitlint` on the
message, `gofmt` + `golangci-lint` on staged Go files (pre-commit), and the race
test suite (pre-push). Never bypass with `--no-verify`.

## Commits & releases

Conventional Commits drive automated versioning via `svu` on GitHub Actions
(Node-free CI/CD; releases published by GoReleaser). Commit type → release effect:

| Type | Effect |
|---|---|
| `feat:` | minor |
| `fix:` | patch |
| `feat!:` / `fix!:` / `BREAKING CHANGE:` | major |
| `chore:`/`docs:`/`refactor:`/`test:`/`ci:`/`build:`/`style:`/`perf:` | no release |

Branches: `develop` is the default (release channel `beta`), `main` is stable.
PRs target `develop`. Both are protected — branch off before committing.

## Skills

Repo-local skills live in `.agents/skills/`:

- `writing-conventional-commits` — group changes and write Conventional Commit
  messages (Go-aware; validates via the Go `commitlint` binary).
- `writing-pull-request-descriptions` — draft PR descriptions targeting `develop`
  using the bundled template.
