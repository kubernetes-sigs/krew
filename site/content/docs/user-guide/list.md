---
title: Listing Installed Plugins
slug: list
weight: 500
---

To list all plugins installed via Krew, run:

```sh
{{<prompt>}}kubectl krew list
```

You can list all installed `kubectl` plugins (including those not installed via
Krew) using:

```sh
{{<prompt>}}kubectl plugin list
```

### Backing up plugin list

When you pipe or redirect the `kubectl krew list` command’s output to another file
or command, it will return a list of plugin names installed, e.g.:

```sh
{{<prompt>}}kubectl krew list | tee backup.txt
access-matrix
whoami
tree
```

You can then [install]({{<ref "install.md">}}) the list of plugins from the file
(on another machine, for example) by feeding the file to the `install` command over
standard input (stdin):

```sh
{{<prompt>}}kubectl krew install < backup.txt
```
