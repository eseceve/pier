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

## Bootstrapping from the cluster

You don't have to write the config from scratch. `pier discover` reads a cluster
with `kubectl` and reports the names you'd otherwise look up by hand — contexts,
namespaces, deployments, ConfigMaps, and Secrets (see
[Usage → discover](usage.md#bootstrapping-the-config-with-discover) for the
flags and JSON shape). It is **read-only**: it prints data, it never writes
`config.yaml`.

How the discovered names map onto the schema above:

- **`contexts`** → the `context` of each `environments` entry.
- **`namespaces`** → a service's `namespace` (or an env `overrides.<env>.namespace`).
- **`deployments`** → the services themselves; each deployment name becomes a
  service (and its `deployment` field when it differs from the service name).
- **`configmaps` / `secrets`** → matched to a service by name: an exact match on
  the service name wins, otherwise a single `"<service>-"`-prefixed candidate is
  used; anything ambiguous is left for you to set explicitly.

When the same service resolves differently across environments (e.g. a different
namespace or Secret on prod), those differences go under `overrides.<env>`.

The **`pier-config`** skill automates this end to end — running discovery across
your contexts and namespaces, proposing the service mapping, and writing the
finished file. Ask it to "build my pier config".
