# pier documentation

`pier` is a small `kubectl` wrapper that maps a friendly service name to its
Kubernetes coordinates (context + namespace, plus optional resource names) via a
config file — **targeting `staging` by default**. Start with the top-level
[`README.md`](../README.md) for the user-facing overview.

## Map

### Available now

- [Installation](installation.md) — `go install` and prebuilt binaries.
- [Usage](usage.md) — command-by-command walkthrough.
- [Configuration](configuration.md) — config file location, schema, and
  resolution.
- [Architecture](architecture.md) — the thin-wrapper design and rationale.
- [Command reference](commands/) — generated per-command reference.
- [Design specs](design/) — the rationale behind `pier` and its subsystems.
- [Implementation plans](plans/) — the phased build plan.

### Coming with later branches

See the
[documentation-infrastructure design](design/2026-07-07-pier-docs-infra-design.md)
for the branch mapping:

- **Agent usage** (`agents.md`) — JSON / `discover` modes and parseable output —
  with `feat/agent-consumable-cli`.

## Generation convention

The per-command reference is **generated** from the cobra command tree, not
written by hand — run `make docs` (which runs `go run ./tools/gendocs` →
`cobra.GenMarkdownTree` → [`commands/`](commands/)). Regenerate it whenever a
command, flag, or help string changes.

Man pages (`cobra/doc.GenManTree`) and shell completions (cobra's built-in
`completion` command for bash / zsh / fish / powershell) follow the same
command tree; see [`commands/README.md`](commands/README.md).

Documentation is plain Markdown rendered natively on GitHub — no static-site
generator and no Node toolchain. Verification is manual review plus GitHub
rendering.
