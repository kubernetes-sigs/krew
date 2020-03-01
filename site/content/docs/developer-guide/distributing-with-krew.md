---
title: Distributing plugins on Krew
weight: 110
---

Once you [develop]({{< ref "develop/plugin-development.md" >}}) a `kubectl`
plugin, here are the steps you need to follow to distribute your plugin on Krew:

1. Package your plugin into an archive file (`.tar.gz` or `.zip`).
1. Make the archive file publicly available (e.g. as GitHub release files).
1. Write [Krew plugin manifest]({{< ref "plugin-manifest.md" >}}) file.
1. [Submit your plugin to [krew-index]({{< ref "release/../release/submitting-to-krew.md" >}}).
