---
title: Automating plugin updates
slug: updating-plugins.md
weight: 300
---

Normally, [releasing a new version]({{< ref "plugin-updates.md" >}}) requires manual
work and creating a pull request every time you have a new version.

However, can use **Github Actions** to publish new release of your Krew plugin.

`krew-release-bot` is a Github Action to automatically bump the version in
`krew-index` repo every time you push a new git tag to your repository:

- It requires no secrets (e.g. GITHUB_TOKEN) to operate.
- It creates your plugin manifest dynamically from a template you write.
- It makes pull requests on your behalf to `krew-index` repository.

Refer to the [krew-release-bot](https://github.com/rajatjindal/krew-release-bot)
documentation for details.

It is **strongly recommended** you automate your plugin's releases. Trivial
version bumps are automatically tested and merged without human intervention,
usually under 5 minutes ([see an example of bots talking to each
other](https://github.com/kubernetes-sigs/krew-index/pull/490)).
