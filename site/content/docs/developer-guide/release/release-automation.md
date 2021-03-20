---
title: Automating plugin updates
slug: automating-updates
weight: 300
---

Normally, [releasing a new version]({{< ref "plugin-updates.md" >}}) requires manual
work and creating a pull request every time you have a new version.

However, you can use a **Github Action** to publish new releases of your Krew plugin.

Specifically, `krew-release-bot` is a Github Action to automatically bump the version in
`krew-index` repo every time you push a new git tag to your repository:

- It requires no secrets (e.g. `GITHUB_TOKEN`) to operate.
- It creates your plugin manifest dynamically from a template you write.
- It makes pull requests on your behalf to the `krew-index` repository.

Refer to the [krew-release-bot](https://github.com/rajatjindal/krew-release-bot)
documentation for details.

The Krew team **strongly recommends** automating your plugin's releases. Trivial
version bumps are automatically tested and merged without human intervention,
usually under five minutes ([see an example of bots talking to each
other](https://github.com/kubernetes-sigs/krew-index/pull/490)).
