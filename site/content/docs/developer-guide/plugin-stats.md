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

1. Visit the [stats.krew.dev] dashboard.
2. Click the “Individual Plugin Stats” report.
3. Choose your plugin from the dropdown and browse the data for your plugin.

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
> [here](https://github.com/corneliusweig/krew-index-tracker).

[GitHub releases]: https://help.github.com/en/github/administering-a-repository/managing-releases-in-a-repository
[stats.krew.dev]: https://datastudio.google.com/c/reporting/f74370a0-adcf-4cec-b7bd-a58c638948f5/page/Ufl7
[GitHub API]: https://developer.github.com/v3/repos/releases/#list-assets-for-a-release
