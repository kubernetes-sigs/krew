---
title: Distributing plugins on Krew
weight: 110
---

## Why Krew? {#why}

To install a kubectl plugin to a user’s machine, all you need to do is to place
an executable in user’s `PATH` prefixed with `kubectl-` name. At this point, you
might consider some other options such as:

- have the user manually download the plugin binary and set up somewhere in
  `PATH`
- distribute the plugin executable using an OS package manager, like Homebrew
  (macOS), apt/yum (Linux) or Chocolatey (Windows)
- distribute the plugin executable using a language package manager (e.g. npm,
  go get)

While these approaches might work some shortcomings to think about are:

- how to ship updates to users (in case of manual installation)
- need to package a plugin for multiple platforms (macOS, Linux, Windows)
- your users need to download the language package manager (go, npm)
- what if you change the implementation language (e.g. move from npm to another package manager)

Krew solves these problems cleanly for all kubectl plugins, since it's designed
**specifically to address these shortcomings**: With Krew, you write a plugin
manifest once and have a plugin that can be installed on all platforms
without having to deal with their package managers.

## Steps to get started

Once you [develop]({{< ref "develop/plugin-development.md" >}}) a `kubectl`
plugin, here are the steps you need to follow to distribute your plugin on Krew:

1. Package your plugin into an archive file (`.tar.gz` or `.zip`).
1. Make the archive file publicly available (e.g. as GitHub release files).
1. Write [Krew plugin manifest]({{< ref "plugin-manifest.md" >}}) file.
1. [Submit your plugin to krew-index]({{< ref "release/../release/submitting-to-krew.md" >}}).
