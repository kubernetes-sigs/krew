---
title: Upgrading Plugins
slug: upgrading-plugins
weight: 600
---

To upgrade all plugins that you have installed to their latest versions, run:

```sh
{{<prompt>}}kubectl krew upgrade
```

Since Krew itself is a plugin also managed through Krew, running the upgrade
command will also upgrade your `krew` setup to the latest version.

To upgrade only certain plugins, you can explicitly specify their names:

```sh
{{<prompt>}}kubectl krew upgrade <PLUGIN1> <PLUGIN2>
```
