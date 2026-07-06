# ADR-0001: Make pier agent-consumable without normalizing kubectl output

**Status:** Accepted
**Date:** 2026-07-06
**Deciders:** Sebas CV (owner), CTO (reviewer)

---

## Context and Problem Statement

`pier` is a small `kubectl` wrapper that maps a friendly service name to its
Kubernetes coordinates (context + namespace + deployment/configmap/secret) via a
config file (`internal/config/config.go`). Its day-to-day commands
(`restart`, `logs`, `status`, `config`, `secret`) are increasingly run not only
by humans but by LLM agents — Claude Code executes `pier ...` over bash, reads
the output, and decides the next step. A `pier-config` Claude skill already
exists (`.claude/skills/pier-config/`) to build the config conversationally.

Running a CLI from an agent has failure modes that don't exist for a human at a
terminal. The two that matter here:

1. **Hangs.** A command that blocks on `stdin` or spawns `$EDITOR` never returns
   for an agent — it stalls the whole session with no way to recover. Today
   `guard()` prompts on `stdin` for mutating ops against a protected env
   (`cmd/root.go:77-82`, via `internal/confirm`), and `config edit` / `secret
   edit` launch `$EDITOR` (`cmd/config_edit.go:9`). Both hang a non-interactive
   caller.

2. **Unparseable output.** Only `discover` emits structured data
   (`discover --json`, `cmd/discover.go:68`). `services` prints a tabwriter table
   (`cmd/services.go:19`); `status`/`logs`/`config get` forward kubectl's raw
   text. An agent parsing tables or free text is brittle, and there is no way to
   ask pier "what are the resolved coordinates for service X in env Y?" without
   executing a side-effecting command — `Config.Resolve` (`config.go:69`) is only
   reachable internally.

`pier` already has several agent-friendly traits worth preserving: `--dry-run`
prints the exact `kubectl` command without running it (`cmd/root.go:65-68`),
errors are self-correcting (`unknown service %q (configured: %s)`,
`config.go:81`), and exit codes from kubectl propagate (`main.go:22-29`).

The decision is **how far to go** in reshaping pier's interface for agents —
specifically, whether pier should own a normalized JSON schema over kubectl's
output, or stay a thin wrapper and expose structure only for the data it owns.

## Decision Drivers

- **No hangs in non-interactive contexts.** A mutating op or an editor command
  invoked without a TTY must fail fast with an actionable error, never block.
- **Deterministic, parseable output for pier-owned data**, so an agent doesn't
  scrape tables or free text.
- **Low maintenance surface.** pier is a thin wrapper maintained by a solo owner;
  it must not take on a schema it has to keep in sync with kubectl.
- **Portability across harnesses.** The interface should help any LLM caller
  (bash-executing agents generally), not only Claude Code.
- **Preserve the human UX.** Interactive prompts and `$EDITOR` must keep working
  at a real terminal.
- **Consistency with kubectl**, which users and agents already know.

## Considered Options

- **Option A:** Thin wrapper — fail-fast on non-TTY, `-o json` for pier-owned
  commands only, delegate structured resource output to kubectl's own `-o`.
- **Option B:** Normalize everything — pier owns a stable JSON schema for all
  commands, including a reshaped view of kubectl's pod/deployment/log output.
- **Option C:** Ship an MCP server exposing pier's commands as typed tools.

## Decision Outcome

**Chosen option: Option A** — pier stays a thin kubectl wrapper. It closes the
hang hazards (fail-fast on non-TTY for both the confirmation prompt and the
editor commands), adds `-o json` to the commands whose data it fully owns, and
introduces a `pier resolve` command that exposes `Config.Resolve` as a
queryable, side-effect-free primitive. For resource data (pods, deployments,
logs), agents use kubectl's native `-o json` rather than a pier-normalized shape.

This wins because it satisfies the "no hangs" and "parseable pier-owned data"
drivers at the lowest maintenance cost, without pier taking ownership of a schema
that mirrors kubectl (the main risk of Option B) or standing up a separate
serving surface (Option C).

### Scope of this decision

- **Fail-fast on non-TTY:** when `stdin` is not a terminal, the confirmation
  prompt is replaced by an error instructing the caller to pass `--yes`, and the
  `edit` commands error out instead of launching `$EDITOR`.
- **`-o/--output` flag** (values `json`, `yaml`) mirroring kubectl, registered
  **per pier-owned command** (`services`, the new `resolve`) rather than as a
  persistent global flag — so a pass-through command errors on an unknown `-o`
  instead of silently ignoring it (a self-correcting signal for agents).
  `discover` keeps its existing `--json` flag untouched: it is already
  structured and the `pier-config` skill depends on it. Pass-through commands
  document that the agent should use kubectl's own `-o`.
- **`pier resolve <service> [-e env] [-o json]`:** prints the fully resolved
  `Target`. With no `-o`, human-readable; with `-o json`, machine-readable.
- **stdout/stderr discipline:** structured output goes only to stdout;
  diagnostics, prompts and `--verbose` traces go to stderr (already the case for
  `--verbose`, `cmd/root.go:70`) — made an explicit invariant.

### Consequences

**Positive:**
- Agents can no longer hang pier; failures are actionable and self-correcting,
  matching the existing error style.
- `pier resolve -o json` + `--dry-run` gives an agent full planning capability
  with zero side effects — it can inspect coordinates and preview the exact
  kubectl command before acting.
- No new schema to maintain against kubectl; the pass-through commands stay as
  thin as they are today.
- The `-o` convention is familiar to anyone who uses kubectl.

**Negative:**
- Structured output is **inconsistent by design**: `pier services -o json` works,
  but `pier status -o json` does not — the agent must fall back to kubectl for
  resource data. This boundary must be documented clearly or it will surprise
  callers.
- Agent-driven writes to ConfigMaps/Secrets remain unsupported: `edit` is
  human-only. Agents can read (`config get`) but not mutate config non-interactively.

**Neutral / requiring follow-up:**
- If a non-Claude-Code consumer ever needs typed tools, revisit Option C (MCP) as
  a layer *on top of* this CLI, not a replacement.
- A general "pier-ops" skill (safe operational workflow: staging default, dry-run
  first, explicit `--yes`) is a natural complement but out of scope here.

---

## Options Analysis

### Option A: Thin wrapper — fail-fast + `-o json` for pier-owned data

pier fixes the hang hazards and emits structured output only where it fully owns
the data (config resolution, the services list, discovery). For everything that
is fundamentally kubectl's output, it stays a pass-through and lets the agent use
`kubectl ... -o json` directly.

**Pros:**
- Directly kills the two hang failure modes.
- Adds the highest-value primitive (`resolve`) by exposing code that already
  exists (`Config.Resolve`).
- Minimal, well-bounded maintenance surface — coherent with pier's stated
  identity as a "small kubectl wrapper" (README).
- Portable to any bash-executing agent, not tied to Claude Code.

**Cons:**
- Deliberately inconsistent structured-output coverage; requires clear docs on
  the pier-owned vs. pass-through boundary.
- Does not enable agent-driven config/secret writes.

---

### Option B: Normalize everything — pier owns a JSON schema for all commands

pier defines and maintains a stable JSON representation for every command,
including reshaped views of kubectl pod/deployment/log output, so an agent gets a
uniform schema regardless of command.

**Pros:**
- Fully consistent output contract across all commands.
- Agents never need to know kubectl exists.

**Cons:**
- pier takes ownership of a schema that mirrors kubectl's output and must track
  it across kubectl versions — brittle and high-maintenance for a solo-owned
  tool (violates the low-maintenance driver).
- Large, speculative surface for uncertain benefit (YAGNI): agents can already
  get structured resource data from `kubectl -o json`.
- Turns a thin wrapper into a translation layer, contradicting pier's identity.

---

### Option C: MCP server exposing pier commands as tools

pier (or a sidecar) runs an MCP server presenting each operation as a typed tool
with a JSON schema, consumed directly by MCP-capable clients.

**Pros:**
- Strongly typed, discoverable interface; no output parsing at all.
- First-class integration for MCP-capable agents.

**Cons:**
- Significant new surface (server lifecycle, transport, auth, tool schemas) for a
  tool that is a thin bash-invokable wrapper — over-engineering for the current
  need.
- Claude Code, the only known consumer, already runs `pier` over bash and is
  covered by a skill; MCP adds no capability the CLI + skill lacks today.
- Best deferred: it can be layered on top of a well-behaved CLI later if a
  non-bash consumer appears, so building it now buys nothing and costs upkeep.

---

## Resolved during implementation

- **`-o` value set:** both `json` and `yaml` (mirrors kubectl; `yaml.v3` is
  already a dependency).
- **`config edit` / `secret edit`:** stay human-only and fail fast without a TTY.
  A future non-interactive `config set` would be a separate ADR.
- **"pier-ops" skill:** out of scope for this cycle; tracked as follow-up.
