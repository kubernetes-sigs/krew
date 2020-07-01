---
title: Hosting Custom Plugin Indexes
slug: custom-indexes
weight: 500
---

Krew comes with a plugin index named `default` that points to the
[`krew-index` repository](https://github.com/kubernetes-sigs/krew) which allows
centralized discovery through community curation.

However, you can host your own plugin indexes (and possibly remove/replace the
`default` index). It’s not recommended to host your own plugin index, unless you
have a use case such as:

- your plugin is not accepted to `krew-index`
- you want full control over the distribution lifecycle of your own plugin
- you want to run a _private_ plugin index in your organization (e.g. to be
  installed on developer machines)

Hosting your own custom index is simple:

- Custom index repositories must be `git` repositories.
- Your clients should have a read access to the repository (if the repository
  is not public, users can still authenticate to it with SSH keys or other
  [gitremote-helpers](https://git-scm.com/docs/gitremote-helpers) installed
  on the client machine).
- The repository must contain a `plugins/` directory at the root, with at least
  one plugin manifest in it. Plugin manifests should be directly in this
  directory.
- Ensure plugins manifests are valid YAML and passes Krew manifest validation
  (optionally, you can use the
  [validate-krew-manifest](https://github.com/kubernetes-sigs/krew/tree/master/cmd/validate-krew-manifest)
  tool for static analysis).

Example plugin repository layout:

```text
.
└── plugins/
    ├── plugin-a.yaml
    ├── plugin-b.yaml
    └── plugin-c.yaml
```

## Duplicate plugin names

Your custom index can contain plugins that have the same name as the ones in
`krew-index`.

Users of your index will need to install your plugin using the
explicit `<INDEX>/<PLUGIN>` syntax. See the [user guide][ug].

[ug]: {{< relref "../user-guide/using-custom-indexes.md">}}
