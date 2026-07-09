# Configuration

`pier` reads a single YAML file that maps each service to its Kubernetes
coordinates. This page covers where that file lives and how a `(service, env)`
pair resolves to a concrete target.

## Location

`pier` resolves the config path in this order:

1. `$PIER_CONFIG`, if set (explicit override).
2. `$XDG_CONFIG_HOME/pier/config.yaml`, if `XDG_CONFIG_HOME` is set.
3. `~/.config/pier/config.yaml` (the fallback).

Check which path is in effect with `pier config path`, and create a starter file
with `pier config init` (it writes to the resolved path and refuses to overwrite
an existing file).

## Schema

```yaml
defaultEnv: staging          # environment used when -e/--env is omitted

environments:
  staging:
    context: staging-cluster # kubectl context name
  prod:
    context: prod-cluster
    protected: true          # mutating ops prompt for confirmation

services:
  api:
    namespace: default
  web:
    namespace: default
    deployment: web-server   # override when the deployment name differs
    configmap: web-config    # optional; defaults to the service name
    secret: web-secret       # optional; defaults to the service name
    overrides:
      prod:
        namespace: web-prod  # this service lives elsewhere on prod
```

### `environments`

Each environment names a `kubectl` `context`. Mark an environment
`protected: true` to require confirmation for mutating operations (`restart`,
`config edit`, `secret edit`) against it.

### `services`

Each service maps its name to Kubernetes resources:

| Field | Meaning | Default when empty |
|---|---|---|
| `namespace` | Namespace for the service | *(no default — resolves empty)* |
| `deployment` | Deployment name | the service name |
| `configmap` | ConfigMap name | the service name |
| `secret` | Secret name | the service name |
| `overrides` | Per-environment field overrides | — |

`overrides.<env>` can change **`namespace`, `configmap`, and `secret`** for a
specific environment — useful when the same service lives in a different
namespace (e.g. `web` → `web-prod` on prod).

## Resolution

Given a service and an environment (`-e/--env`, or `defaultEnv` when omitted),
`pier` resolves each field with this precedence:

```
environment override  →  service field  →  service-name default
```

- **Context** comes from the environment.
- **Namespace** is the env override if set, else the service's `namespace`.
  There is **no** fall back to the service name for namespace.
- **Deployment / ConfigMap / Secret** are the env override if set, else the
  service field, else the service name.
- **Protected** is inherited from the environment.

Resolving an unknown environment or service is an error; the service error lists
the configured names.

You can see the resolved target for any command without running it:

```console
$ pier status web -e prod --dry-run
kubectl --context prod-cluster --namespace web-prod get deployment/web-server
```
