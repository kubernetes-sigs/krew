---
title: Introduction to plugin development
weight: 100
---

If you are looking to start developing plugins for `kubectl`, read the
[Kubernetes
documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)
on this topic.

To summarize the documentation, the procedure is to:

- Write an executable binary, named `kubectl-foo` to have a plugin that can be
  invoked as `kubectl foo`.
- Place the executable to a directory listed in user’s `PATH` environment
  variable. (Plugins distributed with Krew don't need to worry  about this).
- You can't overwrite a builtin `kubectl` command with a plugin.

> **If you are writing a plugin in Go:** Consider using the [cli-runtime] project
> which is designed to provide the same command-line arguments, kubeconfig
> parser, Kubernetes API REST client, and printing logic. Look at
> [sample-cli-plugin] for an example of a kubectl plugin.
>
> Also, there's an unofficial [GitHub template
> repo](https://github.com/replicatedhq/krew-plugin-template) for a Krew plugin
> in Go that implements some best practices mentioned in this guide, and helps
> you automate releases using GoReleaser to create release when a tag is pushed.

[cli-runtime]: https://github.com/kubernetes/cli-runtime/
[sample-cli-plugin]: https://github.com/kubernetes/sample-cli-plugin

While developing a plugin, make sure you check out:

- [Plugin naming guide]({{<ref "naming-guide.md">}}) to choose a good name
  for your plugin
- [Plugin development best practices]({{<ref "best-practices.md">}}) guide
  for a brief checklist of what we're looking for in the submitted plugins.

After you develop a plugin with a good name following the best practices, you
can [develop a Krew plugin manifest]({{<ref "../plugin-manifest.md">}}) and
[submit your plugin to Krew]({{<ref "../release/submitting-to-krew.md">}}).
