## pier discover

Discover cluster resources to help build the config

```
pier discover [flags]
```

### Options

```
      --context string     context to inspect
  -h, --help               help for discover
      --json               emit discovered data as JSON
      --namespace string   namespace to inspect (requires --context)
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

