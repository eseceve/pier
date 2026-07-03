# pier

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`pier` is a small `kubectl` wrapper. It maps a friendly service name to its
Kubernetes coordinates (context + namespace) via a config file, so day-to-day
operations become short and memorable — **targeting staging by default**.

```console
$ pier restart api        # rollout restart on staging
$ pier logs api -f        # follow logs on staging
$ pier restart api -e prod   # prompts for confirmation
```

The name follows the nautical theme of Kubernetes (Greek *kubernetes* =
"helmsman"; the logo is a ship's wheel): a *pier* is where ships dock, mirroring
how each service is "docked" at a known context/namespace.

## Install

Go:

```console
go install github.com/comparaonline/pier@latest
```

Or download a binary from the [releases page](https://github.com/comparaonline/pier/releases).

> A Homebrew tap is planned for a later version.

## Configuration

`pier` reads `$XDG_CONFIG_HOME/pier/config.yaml` (fallback
`~/.config/pier/config.yaml`); override with the `PIER_CONFIG` environment
variable. Generate a starter file with `pier config init`.

```yaml
defaultEnv: staging

environments:
  staging:
    context: staging-cluster
  prod:
    context: prod-cluster
    protected: true            # confirmation required for mutating ops

services:
  api:
    namespace: default
  web:
    namespace: default
    deployment: web-server      # override when the deployment differs
    overrides:
      prod:
        namespace: web-prod
```

## Commands

| Command | Description |
|---|---|
| `pier restart <svc>` | Rollout restart of the service's deployment |
| `pier status <svc>` | Pods + rollout status |
| `pier logs <svc>` | Logs (`-f`, `--tail`, `-c`) |
| `pier config get\|edit <svc>` | Read / edit the service's ConfigMap |
| `pier secret get\|edit <svc>` | Read / edit the service's Secret (`--decode`) |
| `pier services` | List configured services |
| `pier config init\|path` | Create / locate the config file |

Global flags: `-e/--env` (default `staging`), `-y/--yes`, `--dry-run`,
`-v/--verbose`.

## Development

Requires Go 1.26+ and `kubectl` on your `PATH`.

```console
make build      # build ./bin/pier
make test       # go test -race
make lint       # golangci-lint
make install-tools   # golangci-lint, goreleaser, lefthook, svu, commitlint
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the workflow and release process.

## License

[MIT](LICENSE)
