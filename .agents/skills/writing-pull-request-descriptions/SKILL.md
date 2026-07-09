---
name: writing-pull-request-descriptions
description: >-
  Creates and edits pull request descriptions for the pier repo that explain the
  reasoning and architectural decisions behind a change. Writes a Conventional
  Commits title, targets the `develop` base branch, and produces a description
  focused on the "why" rather than the "what" using the repo's default template.
  Use when the user wants to open, create, or draft a pull request — including
  "open a PR", "put up a pull request", "submit this for review", "write the PR
  description", or "my feature is ready to merge". Also triggers when the user says
  "yes", "go ahead", "sounds good", or "ok" and the conversation is about creating
  a PR.
metadata:
  version: 2.0.0
  owner: sebas
  team: dev
  dependencies: 'git, gh CLI'
---

## About this repo

`pier` is a single Go module (`github.com/comparaonline/pier`) — a small `kubectl`
wrapper CLI. There is **no** `commitlint.config.js`; linting runs via the Go
`commitlint` binary against the repo's **`.commitlint.yaml`** (Conventional Commits
defaults plus a `scope-enum`). PR descriptions follow the repo's
**`.github/PULL_REQUEST_TEMPLATE.md`** (Problem / Solution / How to Test). Feature
work targets **`develop`** (the default branch, release channel `beta`); `main` is
the stable channel. Types come straight from Conventional Commits.

## Workflow

### Step 0: Context Scan (before any git commands)

Check what's already in the conversation before running shell commands:

| What to look for | If found | Effect |
|---|---|---|
| User explained the feature/fix this session | Rationale, decisions, alternatives | Use as the primary source for Problem/Solution — skip or minimize `git diff` |
| Commits made this session (e.g. via writing-conventional-commits) | Messages and bodies already in context | Skip `git log`; reuse what's visible |

`git diff origin/develop...HEAD` is the most expensive step. If the user explained
their work, use that for the "why" and use `git log` (lightweight) to verify scope.
Only read the full diff when context is insufficient.

### Step 1: Understand the Context

Gather what's still missing:

```bash
git branch --show-current
git log origin/develop..HEAD --oneline   # skip if commits are already visible in the conversation
git diff origin/develop...HEAD           # skip if the user explained the changes; fallback only
```

**If `git log origin/develop..HEAD` is empty**, the branch has no commits ahead of
`develop` — the work is still uncommitted in the working tree (check
`git status --porcelain`). A PR needs commits, so stop and commit first: hand off to
the `writing-conventional-commits` skill, then resume here. Don't draft a description
from an empty diff.

### Step 2: Craft the PR Title

Format: `<type>(<scope>): <description>`

- **Type**: `feat`, `fix`, `chore`, `refactor`, `docs`, `test`, `ci`, `build`,
  `perf`, `style`. Remember `feat`/`fix` trigger a release via `svu`; pick the
  type that reflects the change's primary intent.
- **Scope** (optional, from the `scope-enum` in `.commitlint.yaml`): the affected
  command (`restart`/`status`/`logs`/`config`/`secret`/`services`) or the Go
  package (`internal/kube` → `kube`, `internal/config` → `config`). Omit if nothing
  in the enum fits (`allow-empty: true`).
- **Description**: brief, imperative ("add retry logic", not "added retry logic"),
  lowercase, no trailing period.

### Step 3: Write the PR Description

Follow the repo's PR template at
[`.github/PULL_REQUEST_TEMPLATE.md`](../../../.github/PULL_REQUEST_TEMPLATE.md) —
its structure is the source of truth. GitHub pre-fills new PRs with it, so the
placeholders are HTML comments; replace them with real prose.

**Problem:** 2-3 sentences on what breaks or is missing. Focus on user/operator
impact, not implementation details.

Good: "`pier restart` against prod doesn't confirm before mutating, risking an
accidental rollout on the wrong cluster."
Weak: "Change the confirmation function."

**Solution (the most important part):** Explain WHY you made the decisions, not just
WHAT changed — reviewers can read the diff. Focus on:

- **Architectural decisions**: why structured this way (e.g. shell out to `kubectl`
  vs. use `client-go`)?
- **Alternatives considered** and why rejected.
- **Trade-offs** and non-obvious choices.

Avoid line-by-line walkthroughs, obvious changes, and repeating commit messages.

**Required sections:** Problem, Solution, How to Test.

**Optional — omit entirely if there's nothing meaningful to say:**
- **Impact / Risks**: breaking changes to config/CLI, behavior against protected
  environments, anything an operator should know before merging.
- **Related**: issue/ticket links, only when they add real context.

### Step 4: Present Draft for Review

Show the user before creating the PR:

```
**Title:** `type(scope): description`

**Description:**
[Full PR body, formatted as it will appear in GitHub]

Does this look good, or would you like me to adjust anything?
```

Wait for confirmation or feedback.

### Step 5: Create the PR

Once approved, push and create the PR against `develop`:

```bash
CURRENT_BRANCH=$(git branch --show-current)
git push -u origin "$CURRENT_BRANCH"
gh pr create --title "title" --body "description" --base develop --assignee @me
```

- Base is `develop` (pier's default branch). Only target `main` when the user
  explicitly says the PR is a stable release/promotion from `develop`.
- `--assignee @me` assigns the PR to the logged-in user.
- `-u` sets up tracking so future pushes are simpler.

`gh` prints the PR URL after creation — share it with the user.

### Edge Cases

- **No commits on the branch:** if the branch has no commits ahead of `develop`
  (uncommitted work only), you can't open a PR yet — commit first (see Step 1).
- **Nothing to push:** `git push` reports "Everything up-to-date" — fine, proceed
  with `gh pr create`.
- **On `develop` (or `main`):** you can't open a PR from the base branch. Ask which
  feature branch to use, or suggest creating one from the changes.
- **PR already exists:** `gh pr create` will report it. Share the existing URL and
  ask whether to update it or branch off.

## Examples

### Example 1: Feature Addition

**Title:** `feat(config): support per-service resource overrides by environment`

```markdown
## Problem

Some services use different Kubernetes coordinates per environment — `web` lives in
namespace `default` on staging but `web-prod` on prod. Today the config only holds
one namespace per service, so `pier ... -e prod` targets the wrong namespace.

## Solution

Added an optional `overrides.<env>` block per service that is merged on top of the
service's base coordinates at resolve time. Chose a merge (base + env override) over
requiring the full coordinates per environment because most services share the same
namespace across environments — the override stays opt-in and keeps configs small.

Considered a flat `service@env` key scheme but rejected it: it duplicates shared
fields and makes the common case (no override) noisier. Resolution precedence is
override → base → service-name default, so unspecified fields still fall back
cleanly.

## How to Test

1. Add an `overrides.prod.namespace: web-prod` block to a service in your config
2. `pier services` — confirm the base (staging) namespace is unchanged
3. `pier status web -e prod --dry-run` — confirm the printed `kubectl` command uses
   `--namespace web-prod`
```

### Example 2: Bug Fix

**Title:** `fix(kube): propagate kubectl's exit code instead of always exiting 0`

```markdown
## Problem

When the underlying `kubectl` call fails (e.g. deployment not found), `pier` still
exits 0. Scripts and CI that check `pier restart`'s exit status treat failures as
successes.

## Solution

Capture the child process's exit code from the shell-out and re-exit with it, rather
than discarding the error after streaming stdio. Kept the passthrough model (we
inherit `kubectl`'s stdio and now its exit code too) instead of parsing `kubectl`
output, which would be brittle and couple `pier` to `kubectl`'s message formats.

## How to Test

1. `pier restart nonexistent-svc` against a reachable cluster
2. Confirm the `kubectl` error is printed and `echo $?` is non-zero (matches what
   running `kubectl` directly would return)

## Impact / Risks

Callers that (incorrectly) relied on `pier` always exiting 0 will now see real
failures surface. No config or command-surface changes.
```

## Important Rules

- **Base branch is `develop`** — don't target `main` unless explicitly promoting a
  stable release.
- **All content in English** — titles, descriptions, everything.
- **Focus on "why" not "what"** — reviewers can read the diff.
- **Keep it concise but complete** — respect reviewer time while giving the context
  needed to understand the decisions.

## Tips

**Think like a reviewer:** what context helps someone understand your decisions?
Address likely questions proactively.

**Future-proof:** six months from now someone may wonder why this works the way it
does — the PR description (preserved in GitHub history) is their answer.
