---
title: Custom Indexes
slug: custom-indexes
weight: 500
---

Custom indexes allow you to distribute your own custom plugins without having to
go through the central `krew-index`. Hosting your own custom index is as simple
as creating a git repository with the following structure:
```
custom-index/
  - plugins/
    - bar.yaml
    - ...
```

Your custom index should contain a `plugins/` directory with at least one plugin
manifest in it.

Users will be able to access your custom index through Krew as long as they're
able to access the repository through git.
