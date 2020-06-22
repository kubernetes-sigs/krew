---
title: Overview
slug: overview
weight: 100
---

Custom indexes allow plugins to be hosted outside of the central index. Users
can then add the custom index to start installing plugins from it.

## Adding a custom index

A custom index can be added with the `kubectl krew index add [name] [git URL]` command
```sh
{{<prompt>}}kubectl krew index add foo https://github.com/foo/custom-index.git
```

## Removing a custom index

You can remove a custom index by passing the name was added with to the remove command
```sh
{{<prompt>}}kubectl krew index remove foo
```

## Listing indexes
To see what indexes you have added you can run `kubectl krew index list`
```sh
{{<prompt>}}kubectl krew index list
{{<output>}}INDEX    URL
default  https://github.com/kubernetes-sigs/krew-index.git
foo      https://github.com/foo/custom-index.git{{</output>}}
```

## Caveats

You can only have *one* plugin with a given name installed at a time. If both
the central krew-index and a custom index you've added contain a plugin named
`grep` then you will only be able to have one installed at a time.
