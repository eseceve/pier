## pier status

Show the service's deployment and rollout status

### Synopsis

Show the service's deployment summary and rollout status. For structured
resource data, append kubectl's own -o (e.g. use kubectl directly), since status
forwards kubectl's output verbatim.

```
pier status <service> [flags]
```

### Examples

```
  pier status api                  # pods + rollout status on staging
  pier status api -e prod
```

### Options

```
  -h, --help   help for status
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

