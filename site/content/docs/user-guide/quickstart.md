---
title: Quickstart
slug: quickstart
weight: 100
---

Krew helps you discover and install [`kubectl` plugins][kpl] on your machine.
There are [a lot of][list] `kubectl` plugins you can install and use to enhance
your Kubernetes experience.

Let's get started.

1. [Install and set up]({{<ref "setup/install.md">}}) Krew.
1. Download the plugin list:

    ```sh
    {{<prompt>}}kubectl krew update
    ```

1. Discover plugins available on Krew:

    ```sh
    {{<prompt>}}kubectl krew search
    {{<output>}}
NAME                            DESCRIPTION                                         INSTALLED
access-matrix                   Show an RBAC access matrix for server resources     no
advise-psp                      Suggests PodSecurityPolicies for cluster.           no
auth-proxy                      Authentication proxy to a pod or service            no
bulk-action                     Do bulk actions on Kubernetes resources.            no
ca-cert                         Print the PEM CA certificate of the current clu...  no{{</output>}}
    ```

1. Install a plugin:

    ```sh
    {{<prompt>}}kubectl krew install access-matrix
    ```

1. Use the installed plugin:

    ```sh
    {{<prompt>}}kubectl access-matrix
    ```

1. Keep your plugins up-to-date:

    ```sh
    {{<prompt>}}kubectl krew upgrade
    ```

1. Uninstall a plugin you no longer use:

    ```sh
    {{<prompt>}}kubectl krew uninstall access-matrix
    ```

This is pretty much all you need to know as a user to use Krew.

[kpl]: https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/
[list]: https://github.com/kubernetes-sigs/krew-index/blob/master/plugins.md
