---
title: Advanced Configuration
slug: advanced-configuration
weight: 900
---

### Use a different default index

When Krew is installed, it automatically initializes an index named `default`
pointing to the [krew-index][ki] repository. You can force Krew to use a
different repository by setting `KREW_DEFAULT_INDEX_URI` before running the
[installation instructions]({{<ref "setup/install.md">}}) or after [removing the
default index]({{<ref "using-custom-indexes.md#the-default-index">}}).
`KREW_DEFAULT_INDEX_URI` should point to a git repository URI that uses a valid
git remote protocol.

To use a different default index, set the `KREW_DEFAULT_INDEX_URI` environment
variable in your `~/.bashrc`, `~/.bash_profile`, or `~/.zshrc`:

```shell
export KREW_DEFAULT_INDEX_URI='git@github.com:foo/custom-index.git'
```

[ki]: https://github.com/kubernetes-sigs/krew-index
