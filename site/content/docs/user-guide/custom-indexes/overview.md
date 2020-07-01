---
title: Overview
slug: overview
weight: 100
---

Custom indexes allow plugins to be hosted outside of `krew-index`. Users can
then add the custom index to start installing plugins from it.

## Adding a custom index

A custom index can be added with the `kubectl krew index add` command
```sh
{{<prompt>}}kubectl krew index add foo https://github.com/foo/custom-index.git
{{<output>}}WARNING: You have added a new index from "https://github.com/foo/custom-index.git"
The plugins in this index are not audited for security by the Krew maintainers.
Install them at your own risk.{{</output>}}
```

The URI you use can be any valid URI (e.g., `git@github.com:foo/custom-index.git`)

## Removing a custom index

You can remove a custom index by passing the name it was added with to the remove command
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

## Caveats

You can only have *one* plugin with a given name installed at a time. If both
the `krew-index` and a custom index you've added contain a plugin named `bar`
then you will only be able to have one installed at a time.
