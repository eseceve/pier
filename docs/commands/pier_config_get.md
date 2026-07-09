## pier config get

Read the service's ConfigMap as YAML

### Synopsis

Read the service's ConfigMap. Output is kubectl's YAML, forwarded verbatim.

```
pier config get <service> [flags]
```

### Examples

```
  pier config get api
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
      --dry-run      print the kubectl command without running it
  -e, --env string   target environment (default: config defaultEnv)
  -v, --verbose      print the kubectl command before running it
  -y, --yes          skip confirmation for mutating ops
```

### SEE ALSO

* [pier config](pier_config.md)	 - Read, edit, locate or create the config / a service's ConfigMap

