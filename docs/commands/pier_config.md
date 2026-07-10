## pier config

Read, edit, locate or create the config / a service's ConfigMap

### Options

```
  -h, --help   help for config
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
* [pier config edit](pier_config_edit.md)	 - Edit the service's ConfigMap in $EDITOR
* [pier config get](pier_config_get.md)	 - Read the service's ConfigMap as YAML
* [pier config init](pier_config_init.md)	 - Write a starter config file
* [pier config path](pier_config_path.md)	 - Print the resolved config file path

