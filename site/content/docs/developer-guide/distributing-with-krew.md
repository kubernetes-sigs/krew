---
title: Distributing plugins on Krew
weight: 110
---

## Why use Krew for distribution? {#why}

On the surface, installing a kubectl plugin seems simple enough -- all you need to do is to place
an executable in the user’s `PATH` prefixed with `kubectl-` -- that you may be
considering some other alternatives to Krew, such as:

- Having the user manually download the plugin executable and move it to some directory in
  the `PATH`
- Distributing the plugin executable using an OS package manager, like Homebrew
  (macOS), apt/yum (Linux), or Chocolatey (Windows)
- Distributing the plugin executable using a language package manager (e.g. npm or
  go get)

While these approaches are not necessarily unworkable, potential drawbacks to consider include:

- How to get updates to users (in the case of manual installation)
- How to package a plugin for multiple platforms (macOS, Linux, and Windows)
- How to ensure your users have the appropriate language package manager (go, npm)
- How to handle a change to the implementation language (e.g. a move from npm to another package manager)

Krew solves these problems cleanly for all kubectl plugins, since it's designed
**specifically to address these shortcomings**. With Krew, after you write a plugin
manifest once your plugin can be installed on all platforms
without having to deal with their package managers.

## Steps to get started

Once you [develop]({{< ref "develop/plugin-development.md" >}}) a `kubectl`
plugin, follow these steps to distribute your plugin on Krew:

1. Package your plugin into an archive file (`.tar.gz` or `.zip`).
1. Make the archive file publicly available (e.g. as GitHub release files).
1. Write a [Krew plugin manifest]({{< ref "plugin-manifest.md" >}}) file.
1. [Submit your plugin to krew-index]({{< ref "release/../release/submitting-to-krew.md" >}}).
