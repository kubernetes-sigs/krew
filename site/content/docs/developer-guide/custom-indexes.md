---
title: Hosting Custom Plugin Indexes
slug: custom-indexes
weight: 500
---

Krew comes with a plugin index named `default` that points to the
[`krew-index` repository](https://github.com/kubernetes-sigs/krew), which allows
centralized discovery through community curation.

However, you can host your own plugin indexes (and possibly remove or replace the
`default` index). Hosting your own plugin index is not recommended, unless you
have a use case that specifically calls for it, such as:

- Your plugin is not accepted to `krew-index`
- You want full control over the distribution lifecycle of your own plugin
- You want to run a _private_ plugin index in your organization (e.g.
  for installations on developer machines)

Hosting your own custom index is simple:

- Custom index repositories must be `git` repositories.
- Your clients should have read access to the repository. If the repository
  is not public, users can still authenticate to it with SSH keys or other
  [gitremote-helpers](https://git-scm.com/docs/gitremote-helpers) installed
  on the client machine.
- The repository must contain a `plugins/` directory at the root, with at least
  one plugin manifest in it. Plugin manifests should be directly in this
  directory (not in a subdirectory).
- Ensure plugin manifests are valid YAML and pass Krew manifest validation
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
