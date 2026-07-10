# Installation

`pier` ships as a single static binary (`CGO_ENABLED=0`) for Linux and macOS on
`amd64` and `arm64`. `kubectl` must be on your `PATH` at runtime — `pier` shells
out to it.

## Go

```console
go install github.com/eseceve/pier@latest
```

Requires Go 1.26+. This installs `pier` into `$(go env GOPATH)/bin`; make sure
that directory is on your `PATH`.

## Prebuilt binaries

Download the archive for your OS/arch from the
[releases page](https://github.com/eseceve/pier/releases), extract it, and move
the `pier` binary onto your `PATH`:

```console
tar -xzf pier_<version>_<os>_<arch>.tar.gz
sudo mv pier /usr/local/bin/
```

Each release publishes a `checksums.txt`; verify your download with
`sha256sum -c checksums.txt` (or `shasum -a 256`).

> Releases are built and published by [GoReleaser](https://goreleaser.com/); a
> Homebrew tap is planned for a later version.

## Verify

```console
pier --version
```
