## pier secret edit

Edit the service's Secret in $EDITOR

### Synopsis

Edit the service's Secret in $EDITOR. Interactive-only: requires a TTY and fails fast for non-interactive callers.

```
pier secret edit <service> [flags]
```

### Options

```
  -h, --help   help for edit
```

### Options inherited from parent commands

```
      --dry-run      print the kubectl command without running it
  -e, --env string   target environment (default: config defaultEnv)
  -v, --verbose      print the kubectl command before running it
  -y, --yes          skip confirmation for mutating ops
```

### SEE ALSO

* [pier secret](pier_secret.md)	 - Read or edit a service's Secret

