# Contributing to pier

Thanks for your interest in improving `pier`.

## Prerequisites

- Go 1.26+
- `kubectl` on your `PATH` (runtime dependency of `pier`)
- Optional dev tools: `make install-tools` installs `golangci-lint`,
  `goreleaser`, and `lefthook`.

## Local workflow

```console
make build   # build ./bin/pier
make test    # go test ./... -race -cover
make lint    # golangci-lint run
make fmt     # gofmt -w .
```

Install the git hooks once with `lefthook install`. They run `gofmt` and
`golangci-lint` before each commit and the test suite before each push.

## Commit messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/).
The commit type drives automated versioning:

- `feat:` → minor release
- `fix:` → patch release
- `feat!:` / `fix!:` / `BREAKING CHANGE:` → major release
- `chore:`, `docs:`, `refactor:`, `test:`, `ci:`, `build:` → no release

## Releases

Releases are automated:

1. Merging Conventional Commits to `main` makes
   [release-please](https://github.com/googleapis/release-please) open (or
   update) a release PR that maintains the version and `CHANGELOG.md`.
2. Merging that PR tags the release, which triggers
   [GoReleaser](https://goreleaser.com/) to build cross-platform binaries,
   publish the GitHub Release, and update the Homebrew tap.

Maintainers only: releasing to the Homebrew tap requires a
`HOMEBREW_TAP_GITHUB_TOKEN` repository secret with write access to the
`eseceve/homebrew-tap` repository.
