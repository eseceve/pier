# Usage

`pier` maps a friendly service name to its Kubernetes coordinates and execs
`kubectl` against the resolved context and namespace. Every command takes a
service name (except `services` and the `config init`/`path` helpers) and
**targets `staging` by default** — pass `-e/--env` to target another
environment.

Set up your config first (see [Configuration](configuration.md)); `pier config
init` writes a starter file.

For the exhaustive, generated flag reference see
[`commands/`](commands/README.md). This page is the task-oriented tour.

## Global flags

These persistent flags work on every command:

| Flag | Default | Effect |
|---|---|---|
| `-e, --env <name>` | config `defaultEnv` | Target environment. |
| `-y, --yes` | `false` | Skip the confirmation prompt on mutating ops. |
| `--dry-run` | `false` | Print the `kubectl` command and exit without running it. |
| `-v, --verbose` | `false` | Print the `kubectl` command (to stderr) before running it. |

`--dry-run` is the safe way to see exactly what `pier` would run:

```console
$ pier restart api --dry-run
kubectl --context staging-cluster --namespace default rollout restart deployment/api
```

## Inspecting

### `pier status <service>`

Shows the deployment and its rollout status — runs `kubectl get deployment/...`
followed by `kubectl rollout status deployment/...`.

```console
$ pier status api
```

### `pier logs <service>`

Streams or prints deployment logs.

| Flag | Effect |
|---|---|
| `-f, --follow` | Follow the log stream. |
| `--tail <n>` | Show only the last `n` lines. |
| `-c, --container <name>` | Target a specific container. |

```console
$ pier logs api -f
$ pier logs api --tail 100 -c sidecar
```

### `pier services`

Lists the configured services with their namespace and deployment, as a table.
Reads the local config only — it does not touch the cluster.

```console
$ pier services
SERVICE  NAMESPACE  DEPLOYMENT
api      default    api
web      default    web-server
```

## Mutating (guarded)

Mutating operations against a `protected` environment prompt for confirmation
before running (see [Safety](#safety) below). `staging` is not protected by
default, so these run without a prompt there.

### `pier restart <service>`

Rollout-restarts the service's deployment (`kubectl rollout restart`).

```console
$ pier restart api              # staging, no prompt
$ pier restart api -e prod      # prod is protected → prompts
$ pier restart api -e prod -y   # skip the prompt
```

## ConfigMaps and Secrets

### `pier config get <service>`

Prints the service's ConfigMap as YAML (`kubectl get configmap/... -o yaml`).

### `pier config edit <service>`

Opens the service's ConfigMap in your `$EDITOR` via `kubectl edit`. Mutating, so
it is guarded on protected environments and needs an interactive terminal.

### `pier secret get <service>`

Prints the service's Secret as YAML. Values stay base64-encoded unless you pass
`--decode`:

```console
$ pier secret get api            # base64 values, as kubectl returns them
$ pier secret get api --decode   # values printed in cleartext
```

> `--decode` writes secret values in cleartext to your terminal — mind shell
> history, scrollback, and screen sharing.

### `pier secret edit <service>`

Opens the service's Secret in your `$EDITOR` via `kubectl edit`. Mutating, so it
is guarded on protected environments and needs an interactive terminal.

## Config file helpers

- `pier config init` — write a starter config file (fails if one already
  exists).
- `pier config path` — print the resolved config file path.

See [Configuration](configuration.md) for the file format and resolution rules.

## Safety

Mutating operations (`restart`, `config edit`, `secret edit`) against an
environment marked `protected: true` prompt for confirmation. The prompt shows
the resolved environment, context, namespace, and the exact `kubectl` command,
then asks you to type the service name to proceed:

```
⚠  PROD — context: prod-cluster, namespace: default
   This will: kubectl --context prod-cluster --namespace default rollout restart deployment/api
   Type the service name to confirm (api):
```

The prompt is skipped when `-y/--yes` is passed, when `--dry-run` is active, or
when the target environment is not protected.

## Falling back to kubectl

`pier` deliberately covers only the common operations above. Anything it does
not model — `exec`-ing into a pod, applying manifests, scaling, port-forwarding
— you run with `kubectl` directly. Use `--dry-run` to grab the resolved
`--context`/`--namespace` if you need a starting point.
