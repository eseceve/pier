# Command reference

These pages are **generated** from pier's cobra command tree — do not edit them
by hand. Regenerate after any change to a command, flag, or help string:

```console
make docs   # go run ./tools/gendocs → cobra.GenMarkdownTree
```

Start at [`pier.md`](pier.md); each page links to its subcommands under
**SEE ALSO**.

- [`pier`](pier.md) — root command and global flags
- [`pier status`](pier_status.md)
- [`pier logs`](pier_logs.md)
- [`pier restart`](pier_restart.md)
- [`pier config`](pier_config.md) — [`get`](pier_config_get.md) · [`edit`](pier_config_edit.md) · [`init`](pier_config_init.md) · [`path`](pier_config_path.md)
- [`pier secret`](pier_secret.md) — [`get`](pier_secret_get.md) · [`edit`](pier_secret_edit.md)
- [`pier services`](pier_services.md)
- [`pier resolve`](pier_resolve.md)
- [`pier discover`](pier_discover.md)

## Man pages and completions

Both derive from the same command tree:

- **Man pages:** `cobra/doc.GenManTree` (roff), generated at package/release time
  rather than committed here.
- **Shell completions:** cobra ships a built-in `completion` command —
  `pier completion bash|zsh|fish|powershell` prints a completion script for your
  shell.
