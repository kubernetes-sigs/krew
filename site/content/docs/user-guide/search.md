---
title: Discovering Plugins
slug: discovering-plugins
weight: 300
---

You can find a list of `kubectl` plugins distributed via Krew [here][list].
However, you can find plugins using the command line as well.

## Search available plugins

First, refresh your local copy of the plugin index:

```sh
{{<prompt>}}kubectl krew update
```

To list all plugins available, run:

```text
{{<prompt>}}kubectl krew search
{{<output>}}
NAME             DESCRIPTION                                         INSTALLED
access-matrix    Show an RBAC access matrix for server resources     no
advise-psp       Suggests PodSecurityPolicies for cluster.           no
auth-proxy       Authentication proxy to a pod or service            no
bulk-action      Do bulk actions on Kubernetes resources.            no
ca-cert          Print the PEM CA certificate of the current clu...  no
...{{</output>}}
```

You can specify search keywords as arguments:

```sh
{{<prompt>}}kubectl krew search pod
{{<output>}}
NAME                DESCRIPTION                                         INSTALLED
evict-pod           Evicts the given pod                                no
pod-dive            Shows a pod's workload tree and info inside a node  no
pod-logs            Display a list of pods to get logs from             no
pod-shell           Display a list of pods to execute a shell in        no
rm-standalone-pods  Remove all pods without owner references            no
support-bundle      Creates support bundles for off-cluster analysis    no{{</output>}}
```

## Learn more about a plugin

To get more information on a plugin, run `kubectl krew info <PLUGIN>`:

```sh
{{<prompt>}} kubectl krew info tree
{{<output>}}
NAME: tree
VERSION: v0.4.0
DESCRIPTION:
  This plugin shows sub-resources of a specified Kubernetes API object in a
  tree view in the command-line. The parent-child relationship is discovered
  using ownerReferences on the child object.
...{{</output>}}
```

[list]: {{< relref "plugins.md" >}}
