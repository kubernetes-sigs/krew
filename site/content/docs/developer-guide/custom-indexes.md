---
title: Hosting Custom Plugin Indexes
slug: custom-indexes
weight: 500
---

[Custom plugin indexes][ug] allow plugin developers to curate and distribute their plugins without
having to go through the centralized [`krew-index`
repository](https://github.com/kubernetes-sigs/krew) (which is a
community-maintained curation of kubectl plugins).

Hosting your own custom index is as simple as creating a git repository with the
following structure:

```text
.
└── plugins/
    ├── plugin-a.yaml
    ├── plugin-b.yaml
    └── plugin-c.yaml
```

- Your custom index should contain a `plugins/` directory with at least one plugin
manifest in it.

- Users will be able to access your custom index through Krew as long as they're
able to access the repository URL through `git`.

## Duplicate plugin names

Your custom index can contain plugins that have the same name as the ones in
`krew-index`.

Users of your index will need to install your plugin using the
explicit `<INDEX>/<PLUGIN>` syntax. See the [user guide][ug].

[ug]: {{< relref "../user-guide/using-custom-indexes.md">}}
