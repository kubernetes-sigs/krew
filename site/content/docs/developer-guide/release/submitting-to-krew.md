---
title: Submitting plugins to Krew
slug: new-plugin
weight: 100
---

Krew maintains a centralized plugin index at the [krew-index] repository.
This repository is cloned on each Krew user’s machine to help them find
existing plugins.

## Pre-submit checklist

1. Review the [Krew plugin naming guide]({{< ref
   "../develop/naming-guide.md" >}}).
1. Read the [plugin development best practices]({{< ref
   "../develop/best-practices.md" >}}).
1. Make sure your plugin’s source code is available as open source.
1. Adopt an open source license, and add it to your plugin archive file.
1. Make sure to extract the LICENSE file during the plugin installation.
1. Tag a git release with a [semantic
   version](https://semver.org) (e.g. `v1.0.0`).
1. [Test your plugin installation locally]({{< ref "../installing-locally.md" >}}).

## Submitting a plugin to krew-index

Once you've run through the checklist above, create a **pull request** to the
[krew-index] repository with your plugin manifest file (e.g. `example.yaml`) to
the `plugins/` directory.

After your pull request is merged, users will be able to find and [install]({{< ref
"../../user-guide/install.md" >}}) your plugin through Krew.

[krew-index]: https://sigs.k8s.io/krew-index
