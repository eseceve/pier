## pier secret get

Read the service's Secret (use --decode for cleartext)

### Synopsis

Read the service's Secret. Output is kubectl's YAML, forwarded verbatim.

```
pier secret get <service> [flags]
```

### Examples

```
  pier secret get api              # base64-encoded values
  pier secret get api --decode     # cleartext values
```

### Options

```
      --decode   print secret values in cleartext
  -h, --help     help for get
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

