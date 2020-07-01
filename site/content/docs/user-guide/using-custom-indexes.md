---
title: Using Custom Plugin Indexes
slug: custom-indexes
weight: 800
---

Plugin indexes contain plugin manifests, which are documents that describe
installation procedure for a plugin. For discovery purposes, Krew comes with a
`default` plugin index,  plugins hosted in the [`krew-index`
repository](https://github.com/kubernetes-sigs/krew-index).

However, some plugin authors may choose to host their own indexes that contain
their curation of kubectl plugins. These are called "custom plugin indexes".

## Adding a custom index

A custom plugin index can be added with the `kubectl krew index add` command:

```sh
{{<prompt>}}kubectl krew index add foo https://github.com/foo/custom-index.git
```

The URI you use can be any [git remote](https://git-scm.com/docs/git-remote)
(e.g., `git@github.com:foo/custom-index.git`).

## Removing a custom index

You can remove a custom plugin index by passing the name it was added with to
the remove command:

```sh
{{<prompt>}}kubectl krew index remove foo
```

## Listing indexes

To see what indexes you have added you can run the list command

```sh
{{<prompt>}}kubectl krew index list
{{<output>}}INDEX    URL
default  https://github.com/kubernetes-sigs/krew-index.git
foo      https://github.com/foo/custom-index.git{{</output>}}
```

## Installing plugins from custom indexes

Commands for managing plugins (e.g. `install`, `upgrade`) work with custom
indexes as well.

By default, Krew prefixes plugins with a `default/` prefix. However, if you have
a plugin installed from a custom index, you need to specify it in format
`INDEX_NAME/PLUGIN_NAME`.

For example, to install a plugin named `bar` from custom index `foo`:

```sh
{{<prompt>}}kubectl krew install foo/bar
```

Similarly:

- To list all plugins (including the ones from custom indexes), run:

    ```sh
    {{<prompt>}}kubectl krew search
    ```

- To remove a plugin, you don't need to specify its index

    ```sh
    {{<prompt>}}kubectl krew uninstall PLUGIN_NAME
    ```

- To get information about a plugin from a custom index

    ```sh
    {{<prompt>}}kubectl krew info INDEX_NAME/PLUGIN_NAME
    ```


> **Caveat:** If two indexes offer a plugin with the same name, only one can
> be installed at any time.
