---
title: Upgrading Plugins
slug: upgrading-plugins
weight: 600
---

Plugins you are using might have newer versions available.

If you want to upgrade all plugins to their latest versions, run:

```sh
{{<prompt>}}kubectl krew upgrade
```

Since Krew itself is a plugin also managed through Krew, running the upgrade
command will also upgrade your `krew` setup to the latest version.

To upgrade only some plugins, you can explicitly specify their name:

```sh
{{<prompt>}}kubectl krew upgrade <PLUGIN1> <PLUGIN2>
```
