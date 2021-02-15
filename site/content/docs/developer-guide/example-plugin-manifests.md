---
title: Example plugin manifests
slug: example-manifests
weight: 200
---

Learning how to [write plugin manifests]({{< ref "plugin-manifest.md" >}})
can be a time-consuming task.

The Krew team encourages you to copy and adapt plugin manifests of [existing
plugins][list]. Since these are already reviewed and approved, your plugin is
likely to be accepted more quickly.

* **Go:**
  - [tree](https://github.com/kubernetes-sigs/krew-index/blob/master/plugins/tree.yaml):
    supports Windows/Linux/macOS, per-OS builds, extracts all files from the
    archive
  - [sort-manifests](https://github.com/kubernetes-sigs/krew-index/blob/master/plugins/sort-manifests.yaml):
    Linux/macOS only, extracting specific files

* **Bash:**
  - [ctx](https://github.com/kubernetes-sigs/krew-index/blob/master/plugins/ctx.yaml):
    Linux/macOS only, downloads GitHub tag tarball, extracts using wildcards

* **Rust:**
  - [view-allocations](https://github.com/kubernetes-sigs/krew-index/blob/master/plugins/view-allocations.yaml):
    Linux/macOS only, implicitly extracts all files from the archive

[list]: https://github.com/kubernetes-sigs/krew-index/tree/master/plugins
