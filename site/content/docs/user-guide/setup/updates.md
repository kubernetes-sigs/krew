---
title: Update checks
slug: update
weight: 300
---

Krew will occasionally check if a new version is available to remind you to
upgrade to a newer version.

This is done by calling the GitHub API, and we do not collect any data from your
machine during this process.

If you want to disable the update checks, set the `KREW_NO_UPGRADE_CHECK`
environment variable. To permanently disable this, add the following to your
`~/.bashrc`, `~/.bash_profile`, or `~/.zshrc`:

```shell
export KREW_NO_UPGRADE_CHECK=1
```
