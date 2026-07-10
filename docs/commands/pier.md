## pier

kubectl wrapper mapping service names to their k8s context/namespace

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
* [pier restart](pier_restart.md)	 - Rollout restart of the service's deployment
* [pier secret](pier_secret.md)	 - Read or edit a service's Secret
* [pier services](pier_services.md)	 - List configured services with their namespace and deployment
* [pier status](pier_status.md)	 - Show the service's deployment and rollout status

