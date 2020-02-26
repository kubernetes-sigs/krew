---
title: Best practices
slug: best-practices
weight: 300
---

This guide lists practices to use while developing `kubectl` plugins. Before
submitting your plugin to Krew, please review these

While Krew project does not enforce any strict guidelines about how a plugin
works, using some of these practices can help your plugin work for more users
and behave more predictably.

### Choosing a language

Most `kubectl` plugins are written in Go or as bash scripts.

If you are planning to write a plugin with Go, check out:

- [client-go]: a Go SDK to work with Kubernetes API and kubeconfig files
- [cli-runtime]: Provides packages to share code with `kubectl` for printing output or [sharing command-line options][cli-opts]
- [sample-cli-plugin]: An example plugin implementation in Go

### Consistency with kubectl

Krew does not try to impose any rules in terms of the shape of your plugin.

However, if you use the command-line options with `kubectl`, you can make your
users’ lives easier.

For example, it is recommended you use the following command line flags used by
`kubectl`:

- `-h`/`--help`
- `-n`/`--namespace`
- `-A`/`--all-namespaces`

Furthermore, by using the [genericclioptions][cli-opts] package (Go), you can
support the global command-line flags listed in `kubectl options` (e.g.
`--kubeconfig`, `--context` and many others) in your plugins.

### Import authentication plugins (Go)

By default, plugins that use [client-go]
cannot authenticate to Kubernetes clusters on many cloud providers. To overcome
this, add the following import to somewhere in your binary:

```go
import _ "k8s.io/client-go/plugin/pkg/client/auth"
```

[cli-runtime]: https://github.com/kubernetes/cli-runtime/
[client-go]: https://godoc.org/k8s.io/client-go
[cli-opts]: https://godoc.org/k8s.io/cli-runtime/pkg/genericclioptions
[sample-cli-plugin]: https://github.com/kubernetes/sample-cli-plugin
