# Using pier from agents and scripts

`pier` is safe to drive non-interactively — from shell scripts or LLM agents.
Per [ADR-0001](adr/ADR-0001-agent-consumable-cli.md), it stays a thin `kubectl`
wrapper: it exposes its **own** resolution as structured data, but it does
**not** parse or normalize `kubectl`'s output. For structured Kubernetes
resource data, use `kubectl`'s own `-o` through pier's pass-through commands.

## Resolve coordinates without side effects

`pier resolve <service>` prints the Kubernetes coordinates a service resolves to
and runs nothing else — the safe way for an agent to plan before acting.

```console
$ pier resolve api -e prod -o json
{
  "service": "api",
  "env": "prod",
  "context": "prod-cluster",
  "namespace": "default",
  "deployment": "api",
  "configmap": "api",
  "secret": "api",
  "protected": true
}
```

`-o yaml` emits the same fields as YAML; omitting `-o` prints a human table.
Check `protected` to decide whether a mutating call will need `--yes`.

## Structured output: what pier owns vs. what kubectl owns

`-o json|yaml` is only offered on the commands whose data pier fully controls:

| Command | `-o json\|yaml` | Shape |
|---|---|---|
| `pier resolve <svc>` | yes | the resolved target (fields above) |
| `pier services` | yes | array of `{ "name", "namespace", "deployment" }` |
| `pier discover --json` | `--json` | discovered resource-name arrays (see [Usage → discover](usage.md#bootstrapping-the-config-with-discover)) |

```console
$ pier services -o json
[
  {
    "name": "api",
    "namespace": "default",
    "deployment": "api"
  }
]
```

An unknown format is rejected (`-o xml` → error: want json or yaml).

**Pass-through commands** (`status`, `logs`, `config get`, `secret get`) stream
`kubectl`'s output untouched. `config get` / `secret get` already emit YAML; for
any other resource or format, run `kubectl` directly using the coordinates from
`pier resolve` rather than expecting pier to reshape the output.

## Streams and exit codes

- **Structured output → stdout.** Diagnostics, the `--verbose` echo, and
  confirmation prompts → **stderr**. An agent can parse stdout cleanly.
- **Exit code is `kubectl`'s.** pier propagates the child process's exit code, so
  a failed operation surfaces as a non-zero status.

## Non-interactive safety

Mutating operations keep their [safety guarantees](usage.md#safety) but fail
fast instead of hanging when there is no terminal:

- **`restart`, `config edit`, `secret edit` against a `protected` env** do not
  prompt when stdin is not a TTY — they fail and point you at `-y/--yes`. Pass
  `--yes` when the agent has already decided the action is safe.
- **`config edit` / `secret edit` are interactive-only** (they open `$EDITOR`),
  so they require a TTY unless you use `--dry-run` to preview the `kubectl`
  command.
- **`--dry-run`** prints the exact `kubectl` command (to stdout) without running
  it — use it to show a plan or to capture coordinates.

```console
# Preview, then act only if resolution looks right:
$ pier resolve api -e prod -o json
$ pier restart api -e prod --dry-run
kubectl --context prod-cluster --namespace default rollout restart deployment/api
$ pier restart api -e prod --yes        # no prompt, no hang
```

## Building config programmatically

`pier discover --json` returns cluster resource names for config bootstrapping,
and the repo-local **`pier-config`** skill drives the whole flow. See
[Usage → discover](usage.md#bootstrapping-the-config-with-discover) and
[Configuration → Bootstrapping from the cluster](configuration.md#bootstrapping-from-the-cluster).
