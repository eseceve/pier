## pier restart

Rollout restart of the service's deployment

```
pier restart <service> [flags]
```

### Examples

```
  pier restart api                 # rollout restart on staging
  pier restart api -e prod         # prompts for confirmation (or --yes)
  pier restart api -e prod --dry-run   # print the kubectl command only
```

### Options

```
  -h, --help   help for restart
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

