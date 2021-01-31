---
title: Quickstart
slug: quickstart
weight: 100
---

Krew helps you discover and install [`kubectl` plugins][kpl] on your machine.

You can install and use [a wide variety of][list] `kubectl` plugins to enhance
your Kubernetes experience.

Let's get started:

1. [Install and set up]({{<ref "setup/install.md">}}) Krew on your machine.

1. Download the plugin list:

    ```sh
    {{<prompt>}}kubectl krew update
    ```

1. Discover plugins available on Krew:

    ```sh
    {{<prompt>}}kubectl krew search
    {{<output>}}NAME                            DESCRIPTION                                         INSTALLED
access-matrix                   Show an RBAC access matrix for server resources     no
advise-psp                      Suggests PodSecurityPolicies for cluster.           no
auth-proxy                      Authentication proxy to a pod or service            no
[...]{{</output>}}
    ```

1. Choose a plugin from the list and install it:

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

This is practically all you need to know to start using Krew.

[kpl]: https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/
[list]: {{< relref "plugins.md" >}}
