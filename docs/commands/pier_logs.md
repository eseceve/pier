## pier logs

Show the service's deployment logs

```
pier logs <service> [flags]
```

### Examples

```
  pier logs api -f                 # follow logs on staging
  pier logs api --tail 100 -c app  # last 100 lines from the app container
```

### Options

```
  -c, --container string   container name
  -f, --follow             stream logs
  -h, --help               help for logs
      --tail int           lines of recent log to show
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

