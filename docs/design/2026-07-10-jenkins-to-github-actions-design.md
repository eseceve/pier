# Migrating CI/CD from Jenkins to GitHub Actions

- **Date:** 2026-07-10
- **Status:** Accepted
- **Scope:** Replace the Jenkins pipeline with GitHub Actions while preserving the
  existing `svu` + GoReleaser release model.

## Problem

`pier` is now a public GitHub repository (`github.com/eseceve/pier`). Its CI/CD
lives in a `Jenkinsfile` that was never run against a real Jenkins (it is marked
`DRAFT`, see `Jenkinsfile:3`) and depends on org-specific infrastructure: a shared
`agent any`, Docker-in-agent (`docker run golang:1.26`), a global
`jenkinsNotification()` library, and a long-lived PAT credential
(`GitHubJenkinsAccessToken`). None of this exists for a personal public repo, and
standing up Jenkins for it is disproportionate.

GitHub Actions is free with unlimited minutes for public repositories and is native
to GitHub, which also makes it the most secure option here:

- The built-in `GITHUB_TOKEN` is minted per-run and expires when the job ends,
  replacing the long-lived `GitHubJenkinsAccessToken` PAT.
- GitHub-hosted runners are ephemeral single-use VMs, versus a persistent shared
  Jenkins agent.
- Any third-party CI (CircleCI, a GitLab mirror) would instead require storing a
  long-lived GitHub PAT in an external system — strictly worse for a public repo.

## Goals

- Preserve the current release behavior exactly (faithful port): `svu` computes the
  version, `develop` publishes `beta` prereleases, `main` publishes stable, and a
  release is only cut when lint + tests + build pass first.
- Keep the automatic `main` → `develop` back-merge.
- No long-lived personal access tokens. Minimal, scoped permissions. Pin all
  third-party actions to a commit SHA.

## Non-goals

- Changing the release model to tag-driven (explicitly rejected — faithful port).
- Homebrew tap automation (still deferred).
- Reworking `.goreleaser.yaml` (reused unchanged).

## Design

### Single workflow, two jobs

`.github/workflows/ci.yml` with two jobs, mirroring the Jenkins stage order
(Test → Release):

- **`test`** — runs on pull requests targeting `develop`/`main` and on pushes to
  `develop`/`main`.
- **`release`** — `needs: test` so a release is never cut on a red build (matching
  Jenkins' in-order stages), guarded by `if:` to run only on **pushes** to
  `develop`/`main`.

One file with two jobs is chosen over two separate workflows because `needs:`
between jobs is native; gating one workflow on another requires `workflow_run`,
which is more fragile.

### `test` job

- `actions/checkout`.
- `actions/setup-go` with `go-version-file: go.mod` (single source of truth for the
  Go version) and module caching.
- `golangci/golangci-lint-action` (installs and caches the linter).
- `go test ./... -race` and `go build ./...`.

This replaces the `docker run golang:1.26` helpers (`Jenkinsfile:62-72`) with native
`setup-go`.

### `release` job

Faithful port of `release()` (`Jenkinsfile:74-121`):

1. `actions/checkout` with `fetch-depth: 0` and `fetch-tags: true` — `svu` needs full
   history and tags.
2. `go install` `svu`; compute the branch-appropriate next tag — `develop` →
   `svu next --prerelease beta`, `main` → `svu next` (stable). Release only if that
   exact tag does not already exist. A `svu next` vs `svu current` string compare
   can't be used: on `develop` the current tag is a beta (`v0.1.0-beta`) while
   `svu next` is the stable target (`v0.1.0`), so they always differ and every push
   would look "ahead". The tag-existence guard is also idempotent — a re-run or the
   back-merge push never re-cuts an existing release.
3. Configure git as `github-actions[bot]`, create and push the tag.
4. Run GoReleaser via the official `goreleaser/goreleaser-action` with the built-in
   `GITHUB_TOKEN`.
5. Safety net: if GoReleaser fails after the tag is pushed, delete the orphaned tag
   so the next run recomputes instead of seeing the tag already exists and skipping
   (matches `Jenkinsfile:105-111`).

### Automatic `main` → `develop` back-merge

`develop` is a protected branch, and the built-in `GITHUB_TOKEN` cannot push to a
protected branch. To keep the back-merge automatic without a long-lived PAT, the
release job mints a **GitHub App installation token** at runtime via
`actions/create-github-app-token`:

- Setup (one-time, by the maintainer):
  1. Create a GitHub App with `Contents: write`, install it on the repo.
  2. Store `APP_ID` and the App private key as repository secrets.
  3. Add the App to the bypass list of the `develop` ruleset.
- At runtime the App token is ephemeral (~1h) and scoped to this repo. It is used
  only for the `git push origin develop` back-merge step.
- The merge itself stays a plain `git merge --no-edit origin/main` (matching
  `Jenkinsfile:118`): a genuine conflict must fail the run for a human to resolve,
  never auto-resolved.

Because the back-merge push uses the App token (not `GITHUB_TOKEN`), it **does**
re-trigger `ci.yml` on `develop`. This is safe: after the back-merge there are no new
version-bumping commits, so the beta tag `svu` computes already exists and the
tag-existence guard makes the `release` job no-op; the only cost is one extra green
`test` run. A `concurrency` group prevents overlapping runs.

### Security

- Top-level `permissions: contents: read` (read-only by default). The `release` job
  elevates to `contents: write` for tags/releases.
- The App token is the only elevated credential and it is short-lived and
  repo-scoped; no PAT is stored.
- All third-party actions (`setup-go`, `golangci-lint-action`, `goreleaser-action`,
  `create-github-app-token`) are pinned to a full commit SHA, so a compromised tag
  cannot inject code into the pipeline.

### Cleanup

- Delete `Jenkinsfile`.
- Update `CONTRIBUTING.md` CI/CD section (`CONTRIBUTING.md:55-64`) and the Jenkins
  references in `docs/design/2026-07-03-pier-design.md` and `AGENTS.md`.
- Drop `jenkinsNotification()` — GitHub natively notifies on workflow failure.

## Alternatives considered

- **Tag-driven releases** (push a tag manually, GoReleaser runs on the tag event):
  simpler and more GitHub-idiomatic, but drops full automation and the `svu`-driven
  version computation. Rejected — the goal is a faithful port.
- **Back-merge via PR** instead of direct push: respects branch protection with just
  `GITHUB_TOKEN` and no App, but the back-merge stops being automatic. Rejected — the
  back-merge must be automatic.
- **Fine-grained PAT** for the back-merge push: simpler setup than a GitHub App, but
  a long-lived secret tied to a personal account. Rejected in favor of ephemeral App
  tokens.
