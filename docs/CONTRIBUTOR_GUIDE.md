# Contributor Guide

This guide is intended for people who want to start working on `krew` itself. If
you intend to write a new plugin, see the [Developer
Guide](https://krew.sigs.k8s.io/docs/developer-guide/) instead.

## Setting up the environment

Krew is a Go module and requires Go 1.27 or newer. You do not need to set up a
`GOPATH`.

```bash
git clone https://github.com/kubernetes-sigs/krew.git
cd krew
go build ./...
```

The `hack` scripts use Bash. On macOS, the system Bash is old enough that some
tools assume a newer version. Install a current Bash with Homebrew if a script
fails on shell features:

```bash
brew install bash
```

`hack/run-lint.sh` installs the pinned versions of `golangci-lint` and `shfmt`
on first use. Install `shellcheck` separately if you want the optional shell
lint step to run:

```bash
brew install shellcheck   # macOS
sudo apt install shellcheck  # Debian/Ubuntu
```

## Code style

Krew follows standard Go formatting and expects imports to be grouped and
sorted. Install `goimports` and format the Go packages:

```bash
go install golang.org/x/tools/cmd/goimports@latest
goimports -local sigs.k8s.io/krew -w cmd internal pkg integration_test
```

The `site/functions` directory is a separate Go module:

```bash
cd site/functions
goimports -local sigs.k8s.io/krew/site/functions -w .
```

Run the linter to check formatting, shell scripts, and Go static analysis:

```bash
hack/run-lint.sh
```

A boilerplate license header is expected in all source files.

_All new code should be covered by tests._

## Compiling

Use `hack/make-binary.sh` to make a binary in `out/bin/` for your current
platform. Use `hack/make-binaries.sh` to build binaries for all supported
platforms, or `hack/make-all.sh` to build release artifacts.

## Running tests

Run the unit tests, race detector, linter, and code-pattern checks with:

```bash
hack/run-tests.sh
```

To run a single check independently, use the scripts in [`hack/`](../hack), or
run the Go tests directly:

```bash
go test -short ./...
go test -short -race ./...
```

The Netlify function in `site/functions` is a separate module:

```bash
cd site/functions
go build -o /dev/null ./...
go test ./...
```

There are also integration tests that cover high-level krew functionality.
Build the `krew` binary first:

```bash
hack/make-binary.sh
hack/run-integration-tests.sh
```

## Testing `krew` in a sandbox

After making changes to krew, check that it behaves as expected without
modifying your normal krew installation. Set `KREW_ROOT` to a temporary
directory:

```bash
mkdir playground
KREW_ROOT="$PWD/playground" out/bin/krew update
```

Any changes krew makes are then applied under `playground/` instead of
`~/.krew`.

### Testing in a docker sandbox

If `KREW_ROOT` does not provide enough isolation, use the Docker sandbox:

```bash
hack/run-in-docker.sh
```
