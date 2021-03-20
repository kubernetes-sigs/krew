---
title: Best practices
slug: best-practices
weight: 300
---

{{< toc >}}

This guide describes practices to use when developing `kubectl` plugins. Before
submitting your plugin to Krew, please review these practices.

While the Krew project team does not strictly enforce guidelines about how a plugin
works, abiding by these practices can help your plugin work for more users
and behave more predictably.

## Choosing a language

Most `kubectl` plugins are written in Go or as bash scripts.

If you are planning to write a plugin with Go, check out:

- [client-go]: a Go SDK to work with Kubernetes API and kubeconfig files
- [cli-runtime]: a set of packages to share code with `kubectl` for printing output or [sharing command-line options][cli-opts]
- [sample-cli-plugin]: an example plugin implementation in Go

## Consistency with kubectl {#kubectl-options}

Krew does not impose any rules in terms of the shape of your plugin.

However, if you support the common command-line options with `kubectl`, you can make your
users’ lives easier.

For example, it is recommended that you support the following command-line flags used by
`kubectl`:

- `-h`/`--help`
- `-n`/`--namespace`
- `-A`/`--all-namespaces`

Furthermore, by using the [genericclioptions][cli-opts] package (Go), you can
support the global command-line flags listed in `kubectl options` (e.g.
`--kubeconfig`, `--context` and many others) in your plugins.

## Import authentication plugins (Go) {#auth-plugins}

By default, plugins that use [client-go]
cannot authenticate to Kubernetes clusters on many cloud providers. To address
this, include the following import in your plugin:

```go
import _ "k8s.io/client-go/plugin/pkg/client/auth"
```

[cli-runtime]: https://github.com/kubernetes/cli-runtime/
[client-go]: https://godoc.org/k8s.io/client-go
[cli-opts]: https://godoc.org/k8s.io/cli-runtime/pkg/genericclioptions
[sample-cli-plugin]: https://github.com/kubernetes/sample-cli-plugin

## Revise usage and help messages with the `kubectl` prefix {#help-messages}

Many users will discover how to use your plugin by calling it without any arguments (which
might trigger a help message), or with `-h`/`--help` arguments.

Therefore, it is helpful to change your usage strings to show the `kubectl ` prefix before the plugin
name. For example:

```text
Usage:
  kubectl popeye [flags]
```

To determine if your executable is running as a `kubectl` plugin, you can look
at `argv[0]`, which will include the `kubectl-` prefix:

- **Go:**

    ```go
    if strings.HasPrefix(filepath.Base(os.Args[0]), "kubectl-") { }
    ```

- **Bash:**

    ```bash
    if [[ "$(basename "$0")" == kubectl-* ]]; then # invoked as plugin
        # ...
    fi
    ```
