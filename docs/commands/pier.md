## pier

kubectl wrapper mapping service names to their k8s context/namespace

### Synopsis

pier maps a friendly service name to its Kubernetes coordinates
(context + namespace + deployment/configmap/secret) via a config file, so
day-to-day operations are short and memorable — targeting staging by default.

Agent/script-friendly contract:
  - --dry-run prints the exact kubectl command without running it.
  - pier resolve -o json reports a service's resolved coordinates (no side effects).
  - services and resolve accept -o json|yaml; pass-through commands (logs,
    status, config get, secret get) delegate structured output to kubectl's own -o.
  - Structured output goes to stdout; diagnostics and prompts go to stderr.
  - Mutating ops against a protected env require a TTY or an explicit --yes;
    non-interactive callers fail fast instead of hanging.

### Options

```
      --dry-run      print the kubectl command without running it
  -e, --env string   target environment (default: config defaultEnv)
  -h, --help         help for pier
  -v, --verbose      print the kubectl command before running it
  -y, --yes          skip confirmation for mutating ops
```

### SEE ALSO

* [pier config](pier_config.md)	 - Read, edit, locate or create the config / a service's ConfigMap
* [pier discover](pier_discover.md)	 - Discover cluster resources to help build the config
* [pier logs](pier_logs.md)	 - Show the service's deployment logs
* [pier resolve](pier_resolve.md)	 - Print the Kubernetes coordinates a service resolves to
* [pier restart](pier_restart.md)	 - Rollout restart of the service's deployment
* [pier secret](pier_secret.md)	 - Read or edit a service's Secret
* [pier services](pier_services.md)	 - List configured services with their namespace and deployment
* [pier status](pier_status.md)	 - Show the service's deployment and rollout status

