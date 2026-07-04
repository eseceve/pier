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
