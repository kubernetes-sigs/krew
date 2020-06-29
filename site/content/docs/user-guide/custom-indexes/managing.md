---
title: Managing Plugins
slug: managing-plugins
weight: 200
---

All of the same functionality for working with plugins exists for plugins from
custom indexes.

## Discovering Plugins

Searching for plugins doesn't change after adding custom indexes. The syntax
remains the same but now results will show up from the custom indexes you added:
```sh
{{<prompt>}}kubectl krew search
{{<output>}}
NAME             DESCRIPTION                                         INSTALLED
access-matrix    Show an RBAC access matrix for server resources     no
advise-psp       Suggests PodSecurityPolicies for cluster.           no
auth-proxy       Authentication proxy to a pod or service            no
bulk-action      Do bulk actions on Kubernetes resources.            no
ca-cert          Print the PEM CA certificate of the current clu...  no
foo/bar          Example plugin from a custom index.                 no
...{{</output>}}
```
To learn more about a plugin from a custom index you can run the same commmand
`kubectl krew info <INDEX>/<PLUGIN>`:
```sh
{{<prompt>}}kubectl krew info foo/bar
{{<output>}}
NAME: bar
INDEX: foo
VERSION: v0.1.0
DESCRIPTION:
  Example plugin from a custom index.{{</output>}}
```

## Installing Plugins

Plugins can be installed from a custom index by prefacing the plugin name with
the name of the index it comes from. For example, if you added an index with the
name `foo`:
```sh
{{<prompt>}}kubectl krew index add foo https://github.com/foo/custom-index.git
{{<output>}}WARNING: You have added a new index from "https://github.com/foo/custom-index.git"
The plugins in this index are not audited for security by the Krew maintainers.
Install them at your own risk.{{</output>}}
```
then you would be able to install plugins from it as `foo/<PLUGIN>`:
```sh
{{<prompt>}}kubectl krew install foo/bar
{{<output>}}
Updated the local copy of plugin index.
Updated the local copy of plugin index "foo".
Installing plugin: bar
Installed plugin: bar
\
 | Use this plugin:
 |      kubectl bar
 | Documentation:
 |      https://github.com/foo/bar
/{{</output>}}
```

The plugin can then be used like any other:
```sh
{{<prompt>}}kubectl bar
```

## Removing Plugins

When you don't need the plugin anymore then you can uninstall it with:
```sh
{{<prompt>}}kubectl krew uninstall bar
```
