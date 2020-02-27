---
title: Releasing plugin updates
slug: updating-plugins
weight: 200
---

When you have a newer version of your plugin, you can update your plugin
manifest at [krew-index] repository to ship this new version to your users.

This manual operation looks like:

1. Update the `version`, `uri` and `sha256` fields of the plugin manifest file.
1. [Test plugin installation locally]({{< ref "../installing-locally.md" >}})
1. Make a pull request to [krew-index] to update the plugin manifest file.

> **Note:** Ideally, the specified `version:` field should match the release tag
of the plugin. (This helps users and maintainers to easily identify which
version of the plugin they have installed.)

If you only change `version`, `uri` and `sha256` fields of your plugin manifest,
your pull request will be automatically approved, tested and merged ([see an
example](https://github.com/kubernetes-sigs/krew-index/pull/508)).

However, you can [**automate releasing your plugin**]({{< ref
"release-automation.md" >}}) if you're publishing your plugins on GitHub.

[krew-index]: https://sigs.k8s.io/krew-index
