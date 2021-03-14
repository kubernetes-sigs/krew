---
title: Introduction to plugin development
weight: 100
---

If you are looking to start developing plugins for `kubectl`, read the
[Kubernetes
documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)
on this topic.

To summarize the documentation, the procedure is to:

- Create an executable binary, named `kubectl-foo` (for example) to have a plugin that can be
  invoked as `kubectl foo`.
- Place the executable in a directory that is listed in the user’s `PATH` environment
  variable. (You don't have to do this for plugins distributed with Krew).
- You can't overwrite a built-in `kubectl` command with a plugin.

> **Note:** If you are writing a plugin in Go, consider using the [cli-runtime] project,
> which is designed to provide the same command-line arguments, kubeconfig
> parser, Kubernetes API REST client, and printing logic. Look at
> [sample-cli-plugin] for an example of a kubectl plugin.
>
> Also, see the unofficial [GitHub template
> repo](https://github.com/replicatedhq/krew-plugin-template) for a Krew plugin
> in Go that implements some best practices covered later in this guide, and helps
> you automate releases using GoReleaser to create a release when a tag is pushed.

[cli-runtime]: https://github.com/kubernetes/cli-runtime/
[sample-cli-plugin]: https://github.com/kubernetes/sample-cli-plugin

When developing your own plugins, make sure you check out:

- [Plugin naming guide]({{<ref "naming-guide.md">}}) to choose a good name
  for your plugin
- [Plugin development best practices]({{<ref "best-practices.md">}}) guide
  for a brief checklist of what we're looking for in the submitted plugins.

After you develop a plugin with a good name following the best practices, you
can [develop a Krew plugin manifest]({{<ref "../plugin-manifest.md">}}) and
[submit your plugin to Krew]({{<ref "../release/submitting-to-krew.md">}}).
