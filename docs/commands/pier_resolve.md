## pier resolve

Print the Kubernetes coordinates a service resolves to

### Synopsis

Resolve a service name (and environment) to its concrete Kubernetes
coordinates without running anything, so agents and scripts can plan before
acting. Combine with --dry-run on other commands to preview the exact kubectl
invocation.

```
pier resolve <service> [flags]
```

### Examples

```
  pier resolve api                 # resolved coordinates on the default env
  pier resolve api -e prod         # against the prod env
  pier resolve api -o json         # machine-readable, for agents/scripts
```

### Options

```
  -h, --help            help for resolve
  -o, --output string   structured output format: json or yaml
```

### Options inherited from parent commands

```
      --dry-run      print the kubectl command without running it
  -e, --env string   target environment (default: config defaultEnv)
  -v, --verbose      print the kubectl command before running it
  -y, --yes          skip confirmation for mutating ops
```

### SEE ALSO

* [pier](pier.md)	 - kubectl wrapper mapping service names to their k8s context/namespace

