---
title: Advanced Configuration
slug: advanced-configuration
weight: 900
---

{{< toc >}}

## Customize installation directory {#custom-install-dir}

By default, Krew installs itself and plugins to `$HOME/.krew`. This means
Krew itself and the installed plugins will be visible only to the user who
installed it.

To customize this installation path, set the `KREW_ROOT` environment variable
while [installing Krew]({{< relref "setup/install.md" >}}). After Krew is
installed, you still need to set `KREW_ROOT` in your environment for Krew
to be able to find its installation directory.

For example, add this to your `~/.bashrc` or `~/.zshrc` file:

```shell
export KREW_ROOT="/usr/local/krew"
```

Note that you still need to add `$KREW_ROOT/bin` to your `PATH` variable
for `kubectl` to be able to find installed plugins.

## Use a different default index {#custom-default-index}

When Krew is installed, it automatically initializes an index named `default`
pointing to the [krew-index][ki] repository. You can force Krew to use a
different repository by setting `KREW_DEFAULT_INDEX_URI` before running the
[installation instructions]({{<ref "setup/install.md">}}) or after [removing the
default index]({{<ref "using-custom-indexes.md#the-default-index">}}).
`KREW_DEFAULT_INDEX_URI` must point to a git repository URI that uses a valid
git remote protocol.

To use a different default index, set the `KREW_DEFAULT_INDEX_URI` environment
variable in your `~/.bashrc`, `~/.bash_profile`, or `~/.zshrc`:

```shell
export KREW_DEFAULT_INDEX_URI='git@github.com:foo/custom-index.git'
```

## Configure network proxy {#custom-network-proxy}

If you want to use Krew with an HTTP proxy, you can configure environment 
variables HTTP_PROXY, HTTPS_PROXY and NO_PROXY. Behavior of these 
environment variables are explained [here][httpproxy].

```shell
export HTTP_PROXY="proxy-ip:port"
export HTTPS_PROXY="proxy-ip:port"
export NO_PROXY="ip1,ip2:port2,.example.com"
```

[ki]: https://github.com/kubernetes-sigs/krew-index
[httpproxy]: https://pkg.go.dev/golang.org/x/net/http/httpproxy#Config
