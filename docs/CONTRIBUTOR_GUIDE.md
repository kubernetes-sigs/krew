# Contributor Guide

This guide is intended for people who want to start working on `krew` itself.
If you intend to write a new plugin, see [Developer Guide](./DEVELOPER_GUIDE.md)
instead.

## Setting up the environment

Krew is built with go 1.10, but newer versions will do as well.
Most toolchains will expect that the krew repository is on the `GOPATH`.
To set it up correctly, do

```bash
mkdir -p $(go env GOPATH)/src/sigs.k8s.io/krew
cd $(go env GOPATH)/src/sigs.k8s.io/krew
git clone https://github.com/kubernetes-sigs/krew .
git remote set-url origin --push no_push   # to avoid pushes
```

## Code style

Krew adheres to standard `golang` code formatting conventions, and also expects
imports sorted properly.
To automatically format code appropriately, install
[`goimports`](https://godoc.org/golang.org/x/tools/cmd/goimports) via 

```bash
go get golang.org/x/tools/cmd/goimports
```

and run:

```bash
goimports -w cmd pkg
```

In addition, a boilerplate license header is expected in all source files.

_All new code should be covered by tests._


## Running tests

To run tests locally, the easiest way to get started is with

```bash
hack/run-tests.sh
```

This will run all unit tests and code quality tools.
To run a single tool independently of the other code checks, have a look at the
other scripts in [`hack/`](../hack).

In addition, there are integration tests to cover high-level krew functionality.
To run integration tests, you will need to build the `krew` binary beforehand:

```bash
hack/make-binaries.sh
hack/run-integration-tests.sh
```

The above builds binaries for all supported platforms.
Building only for your platform produces a slightly different image but will
work in most circumstances:

```bash
go build ./cmd/krew
# you need to specify your local krew binary when using this method:
hack/run-integration-tests.sh ./krew
```

## Testing `krew` in a sandbox

After making changes to krew, you should also check that it behaves as expected.
You can do this without messing up the krew installation on the host system by
setting the `KREW_ROOT` environment variable.
For example:

```bash
mkdir playground
KREW_ROOT="$PWD/playground" krew update
```

Any changes that krew is going to apply will then be applied in the
`playground/` folder, instead of the standard `~/.krew` folder.

### Testing in a docker sandbox

Alternatively, if the isolation provided by `KREW_ROOT` is not enough, there is
also a script to run krew in a docker sandbox:

```bash
hack/run-in-docker.sh
```
