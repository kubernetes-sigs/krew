---
title: Configuration
slug: configuration
weight: 300
---

### Disable update checks

Krew will occasionally check if a new version is available to remind you to
upgrade to a newer version.

This is done by calling the GitHub API, and we do not collect any data from your
machine during this process.

If you want to disable the update checks, set the `KREW_NO_UPGRADE_CHECK`
environment variable. To permanently disable this, add the following to your
`~/.bashrc`, `~/.bash_profile`, or `~/.zshrc`:

```shell
export KREW_NO_UPGRADE_CHECK=1
```

### Use a different default index

By default Krew uses [krew-index][ki] as the default plugin index. This is the
repository that is normally used for plugin discovery. However, you can use a
different default index if you don't want to use krew-index or if you would
like to use your own release of Krew.

To use a different default index, set the `KREW_DEFAULT_INDEX_URI` environment
variable in your `~/.bashrc`, `~/.bash_profile`, or `~/.zshrc` before running
the [installation instructions]({{<ref "install.md">}}):

```shell
export KREW_DEFAULT_INDEX_URI='git@github.com:foo/custom-index.git'
```

[ki]: https://github.com/kubernetes-sigs/krew-index
