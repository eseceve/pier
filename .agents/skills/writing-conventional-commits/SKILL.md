---
name: writing-conventional-commits
description: >-
  Creates well-structured git commits following Conventional Commits format for
  the pier repo (a Go module). Detects scopes from Go package/command paths,
  groups changes by logical unit (implementation and its tests together),
  validates with the Go commitlint binary when available, and presents a plan
  for approval before executing. Use this skill whenever the user wants to create
  commits, mentions committing changes, or says things like "commit this", "make
  a commit", "ok commit", "save this", or "done, commit it". Also use when the
  user is on a protected branch (`main`/`develop`) and needs a feature branch
  first. Do not invoke for git history, merge conflicts, rebase, squash, or branch
  operations that don't involve creating a new commit.
metadata:
  version: 2.0.0
  owner: sebas
  team: dev
  dependencies: >
    Requires git. Optional (installed via `make install-tools`):

    - commitlint (Go binary, github.com/conventionalcommit/commitlint): validates
    the commit message. Falls back to manual Conventional Commits checks if absent.
---

## About this repo

`pier` is a single Go module (`github.com/eseceve/pier`) — a small `kubectl`
wrapper CLI. There is no Node.js or Python tooling: commit-message linting runs
through the **Go** `commitlint` binary, wired via lefthook (`lefthook.yml`), not
through `@commitlint/cli`. The repo ships a **`.commitlint.yaml`** at the root: it
keeps the Conventional Commits defaults (types, 72-char header) and adds a
**`scope-enum`** that constrains scopes to pier's commands and packages —
`restart`, `status`, `logs`, `config`, `secret`, `services`, `kube` — with
`allow-empty: true`, so a scopeless message is still valid. A curated scope list
keeps the GoReleaser changelog consistent (see `.goreleaser.yaml` groups).

Commit types drive automated versioning via `svu` in Jenkins (see `CONTRIBUTING.md`):

| Type | Release effect |
|---|---|
| `feat:` | minor release |
| `fix:` | patch release |
| `feat!:` / `fix!:` / `BREAKING CHANGE:` footer | major release |
| `chore:`, `docs:`, `refactor:`, `test:`, `ci:`, `build:`, `style:`, `perf:` | no release |

This is why `feat`/`fix` dominate type selection (step 5) — mislabeling a
release-worthy change as `chore` silently skips a release.

## Core Workflow

### 0. Context Scan (before any git commands)

Read the conversation history first and treat it as a cache:

| What to look for | If found | Effect |
|---|---|---|
| Explicit file list ("commit config.go and its tests") | Set file list | Skip `git status` for discovery; still verify files are tracked |
| Rationale ("I replaced X with Y because Z") | Capture verbatim | Skip `git diff` reads for body composition |
| Change type clue ("I fixed…", "I added…", "refactored…") | Pre-select type | Skip inferring type from the diff |

Only fall back to git commands when the context is missing or ambiguous.

### 1. Branch Protection Check

```bash
git branch --show-current
```

If on `main` or `develop` (pier's protected branches — `develop` is the default,
`main` is stable): block, suggest a feature branch named `<type>/<short-description>`
(e.g. `feat/config-discover`, `fix/kubectl-exit-code`), offer to create it with
`git checkout -b <name>`, and wait for the user's decision.

### 2. Scenario Detection

| Scenario | Condition | Action |
|---|---|---|
| User-specified files | File names in message | Process only those files |
| Staged only | `git diff --cached --name-status` has results | Process staged files |
| Complete analysis | `git status --porcelain` shows changes | Analyze and group all changes |
| No changes | Nothing to commit | Inform user, stop |
| Merge conflicts | `git diff --name-only --diff-filter=U` returns files | Block until resolved |
| Explicit amend | User says "amend" | Use `git commit --amend` |

### 3. Change Analysis & Grouping

```bash
git diff --cached --name-status  # Staged
git status --porcelain           # All changes
```

**Renamed files:** `--name-status` outputs renames as `R<similarity>\t<old>\t<new>`.
Use the new path for scope detection; the old path's deletion is already tracked.

**Grouping rule — by logical unit, not by type:** Each commit should represent one
coherent change. A feature and its tests are one commit. A fix and the refactoring
it required are one commit. Ask: "do these files change together to achieve one
goal?" If yes, they belong in the same commit.

Examples (pier):
- Adding `pier config discover` + its `internal/config` test + wiring it into the
  root command → **1 commit**
- Fixing kubectl exit-code propagation + refactoring the exec helper it lived in
  → **1 commit**
- Two unrelated changes (new `--dry-run` flag + a docs fix) → **2 commits**

**Do NOT split by file type.** Tests, config, docs, and implementation that serve
the same logical change go together.

### 4. Commit Message Composition

Format: `type(scope): subject`

**Types:** `feat`, `fix`, `refactor`, `test`, `docs`, `style`, `chore`, `ci`,
`build`, `perf`, `revert`. For breaking changes append `!` to the type and add a
`BREAKING CHANGE:` footer.

**Type selection — release-triggering types dominate.** When a group mixes types,
pick the one representing the primary intent:

| Files in the group | Resulting type |
|---|---|
| New feature + its tests | `feat` |
| Bug fix + refactored surrounding code + tests | `fix` |
| Only test files (no behavior change) | `test` |
| Only refactoring (no new behavior, no bug fix) | `refactor` |
| Only docs | `docs` |

If the group contains any `feat`- or `fix`-like change alongside
`refactor`/`test`/`docs`/`chore`, the type is `feat` or `fix` — because those
trigger a release via `svu`.

Pick the type from what the diff *actually does*, not the file names. A new
`*_test.go` file next to a stub with no behavior change is `test`, not `feat` —
labeling a no-op stub `feat` would trigger a bogus release. The same applies to
the feature-branch name in step 1: its `<type>` should reflect the real change.

**Scope detection — pick from the `scope-enum` in `.commitlint.yaml`:**

The allowed scopes are `restart`, `status`, `logs`, `config`, `secret`,
`services`, `kube`. A scope outside this list is rejected by the `commit-msg`
hook. To map a change to one:

1. **Command name** — a change under the `restart`/`status`/`logs`/`config`/
   `secret`/`services` command uses that command as scope
   (e.g. `feat(logs): add --since flag`).
2. **Package/directory** — strip `cmd/`, `internal/`, `pkg/` and take the first
   meaningful segment (`internal/config/resolve.go` → `config`;
   `internal/kube/exec.go` → `kube`).
3. **Cross-cutting / tooling** — changes with no domain scope (`*.md`, `.github/`,
   `.goreleaser.yaml`, `lefthook.yml`, `Jenkinsfile`, `Makefile`, `go.mod`) go
   **scopeless**, carried by the type alone (`docs:`, `ci:`, `build:`, `chore:`).
4. **Fallback** — omit the scope if nothing in the enum fits; `allow-empty: true`
   keeps a scopeless message valid.

If a new command or package is added, extend the `scope-enum` in
`.commitlint.yaml` in the same change before using its scope.

**Commit body** — include when the change is non-trivial; explain the *why*, not
the *what*. Prefer the conversation as the source; only run `git diff HEAD <file>`
when the context doesn't explain the reasoning.

```
feat(config): resolve per-service namespace overrides by environment

- resolve.go: merge the service's env-specific overrides on top of its base
  coordinates so `-e prod` picks up web-prod without duplicating config
- resolve_test.go: cover the override-precedence and missing-env cases
```

Omit the body only for trivial, self-explanatory changes.

### 5. Validate the Message

The Go `commitlint` binary is optional and may not be installed locally
(`make install-tools` installs it). Check first:

```bash
command -v commitlint >/dev/null && commitlint lint --message "type(scope): subject"
```

- If available: fix any reported error before presenting the plan.
- If absent: validate manually against Conventional Commits — `type(scope): subject`
  in lowercase, imperative subject, no trailing period, header kept concise
  (~72 chars). lefthook's `commit-msg` hook will run `commitlint` at commit time
  regardless, so a well-formed message passes there too.

### 6. Present Plan

```
Branch: feat/config-discover

Proposed Commits:

1. feat(config): resolve per-service namespace overrides by environment
   Files: internal/config/resolve.go, internal/config/resolve_test.go
   Body:
   - resolve.go: merge env-specific overrides on top of base coordinates
   - resolve_test.go: cover override precedence and missing-env cases

Proceed?
```

Wait for approval before executing.

### 7. Execute (if approved)

Stage files only when needed:
- **Unstaged** (user-specified or complete analysis): `git add <files>` first
- **Already staged**: skip `git add` to avoid capturing unstaged changes

Use a heredoc for multi-line bodies:

```bash
git commit -F - << 'EOF'
type(scope): subject

body line 1
body line 2
EOF
```

For breaking changes, put the footer after a blank line following the body:

```bash
git commit -F - << 'EOF'
feat(config)!: subject

body

BREAKING CHANGE: description
EOF
```

**pier's git hooks run automatically** (`lefthook.yml`): `commit-msg` lints the
message, and `pre-commit` runs `gofmt` + `golangci-lint` on staged `*.go` files.
If a commit is rejected, the message or the code needs fixing — never bypass with
`--no-verify`, and never `git add -A`. Verify with `git status` when done.
