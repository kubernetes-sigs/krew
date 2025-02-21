---
title: Plugin usage analytics
slug: plugin-stats
weight: 600
---

Krew does not track user behavior. However, for plugins that distribute their
packages as assets on [GitHub
releases],
Krew offers some statistics to track download analytics for your plugins over
time.

To see your plugin’s download statistics over time:

1. Visit the [krew-index-tracker] dashboard.
2. Choose your plugin from the dropdown and browse the data for your plugin.

This data is obtained by scraping the downloads count of your plugin assets via
the [GitHub API] regularly. Since Krew does not track its users, this data:

- does not reflect active installations
- does not distinguish between installs, reinstalls, and upgrades
- is purely a tracking of download counts of your release assets over time

> **Note:** The Krew plugin stats dashboard is provided as a best effort by Krew
> maintainers to measure the success of Krew and its plugins. We cannot guarantee
> its availability and accuracy.
>
> The scraping code can be found
> [here](https://github.com/predatorray/krew-index-tracker).

[GitHub releases]: https://help.github.com/en/github/administering-a-repository/managing-releases-in-a-repository
[krew-index-tracker]: https://predatorray.github.io/krew-index-tracker/
[GitHub API]: https://developer.github.com/v3/repos/releases/#list-assets-for-a-release
