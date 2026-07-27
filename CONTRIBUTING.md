# Contributing to pier

Thanks for your interest in improving `pier`.

## Prerequisites

- Go 1.26+
- `kubectl` on your `PATH` (runtime dependency of `pier`)
- Optional dev tools: `make install-tools` installs `golangci-lint`,
  `goreleaser`, `lefthook`, `svu`, and `commitlint` (all Go binaries — no Node
  required).

## Local workflow

```console
make build   # build ./bin/pier
make test    # go test ./... -race -cover
make lint    # golangci-lint run
make fmt     # gofmt -w .
```

Install the git hooks once with `lefthook install`. They run `commitlint` on the
commit message, `gofmt` + `golangci-lint` before each commit, and the test suite
before each push.

## Commit messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/).
The commit type drives automated versioning:

- `feat:` → minor release
- `fix:` → patch release
- `feat!:` / `fix!:` / `BREAKING CHANGE:` → major release
- `chore:`, `docs:`, `refactor:`, `test:`, `ci:`, `build:` → no release

## Documentation

Documentation is plain Markdown under [`docs/`](docs/README.md), rendered
natively on GitHub — no static-site generator and no Node toolchain (a Markdown
linter would pull in Node and is deliberately avoided). Verify changes by
reviewing the GitHub-rendered output.

The command reference, man pages, and shell completions are **generated** from
the cobra command tree, not hand-written:

- **Command reference:** `cobra.GenMarkdownTree` → `docs/commands/`.
- **Man pages:** `github.com/spf13/cobra/doc.GenManTree`.
- **Shell completions:** cobra's built-in `completion` command.

The `make docs` / `make man` targets that drive this generation land on the
branch that carries the command tree (`feat/cli-implementation`).

## Releases

CI/CD runs on GitHub Actions ([`.github/workflows/ci.yml`](.github/workflows/ci.yml))
and is fully Node-free. Releases are automated from Conventional Commits:

1. On `develop` (channel `beta`) and `main` (stable), the workflow computes the
   next version with [`svu`](https://github.com/caarlos0/svu) and creates + pushes
   the git tag.
2. [GoReleaser](https://goreleaser.com/) cross-compiles the binaries and
   publishes the GitHub Release (with notes from the commit history).

Homebrew tap automation is planned for a later version.
