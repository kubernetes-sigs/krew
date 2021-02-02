---
title: Installing Plugins
slug: installing-plugins
weight: 400
---

Plugins can be installed with the `kubectl krew install` command:

```text
{{<prompt>}}kubectl krew install ca-cert
{{<output>}}Installing plugin: ca-cert
Installed plugin: ca-cert{{</output>}}
```

This command downloads the plugin and verifies the integrity of the downloaded
file.

After installing a plugin, you can start using it by running `kubectl <PLUGIN_NAME>`:

```sh
{{<prompt>}}kubectl ca-cert
```


