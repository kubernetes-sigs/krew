---
title: Submitting plugins to Krew
slug: new-plugin
weight: 100
---

Krew maintains a centralized plugin index at the [krew-index] repository.
Since this repository is cloned on Krew user’s machine, it helps them find
existing plugins.

## Pre-submit checklist

1. Familiarize yourself with [Krew plugin naming guide]({{< ref "../develop/naming-guide.md" >}}).
1. Read the [plugin development best practices]({{< ref
   "../develop/best-practices.md" >}}).
1. Develop a [plugin manifest]({{< ref "../plugin-manifest.md" >}})
1. Make sure your plugin’s source code is available as open source.
1. Adopt an open source license, and add it to your plugin archive file.
1. Tag a git release with a [semantic
   version](https://semver.org) (e.g. `v1.0.0`)
1. Make sure your plugin archive file (`.tar.gz` or `.zip`) is available
   **publicly**.
1. [Test plugin installation locally]({{< ref "../installing-locally.md" >}}).

## Submitting a plugin to krew-index

Create a **pull request** to the [krew-index] repository with your plugin
manifest (e.g. `example.yaml`) file to the `plugins/` directory.

After your pull request is merged, users can
[install](../../user-guide/install.md) your plugin through Krew.

[krew-index]: https://sigs.k8s.io/krew-index
