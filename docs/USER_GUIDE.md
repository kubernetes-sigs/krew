# User Guide

This guide shows how to use `krew` as a user after installing it.

## Discovering Plugins

To find plugins, run the `kubectl krew search` command. This command lists all
available plugins:

```text
$ kubectl krew search
NAME               SHORT                                              STATUS
ca-cert            Print PEM CA certificate of current cluster        available
extract-context    Extract current-context on kubectl as a kubecon... available
krew               Install plugins                                    available
mtail              Tail logs from multiple pods matching label sel... available
view-secret        Decode secrets                                     available
...
```

You can specify search keywords as arguments:

```text
$ kubectl krew search crt
NAME               SHORT                                       STATUS
ca-cert            Print PEM CA certificate of current cluster available
view-secret        Decode secrets                              available
```

To get more information on a plugin, run `kubectl krew info <PLUGIN>`:

```text
$ kubectl krew info ca-cert
NAME: ca-cert
HEAD: https://github.com/ahmetb/kubectl-extras/archive/master.zip
URI: https://github.com/ahmetb/kubectl-extras/archive/c403c57.zip
SHA256: 8be8ed348d02285abc46bbf7a4cc83da0ee9d54dc2c5bf86a7b64947811b843c
DESCRIPTION:
 Pretty print the current cluster certificate.
 The plugin formats the certificate in PEM following RFC1421.
VERSION: master
CAVEATS:
 This plugin needs the following programs:
 * base64
```

## Installing Plugins

Plugins can be installed with `kubectl krew install` command:

```text
$ kubectl krew install ca-cert

Will install plugin: ca-cert
Installing plugin: ca-cert
Installed plugin: ca-cert
CAVEATS:
 This plugin needs the following programs:
 * base64
```

This command downloads the plugin and verifies the integrity of the downloaded
file.

After installing a plugin, you can use it like `kubectl <PLUGIN>`:

```
$ kubectl ca-cert
```

### Installing plugins with --HEAD

Some plugins are offer a way to install directly from the last revision of the
source code from their Git repositories. Such plugins expose a `HEAD:` field in
`kubectl info` output.

To install such a plugin from its latest release, run:

    kubectl krew install --HEAD <PLUGIN>

**Note:** Installing with `--HEAD` does not check the integrity of the
downloaded git archive. Also, untagged plugins are very likely to be unstable.

## Listing Installed Plugins

All plugins available to `kubectl` (including those not installed via krew) can
listed using:

    kubectl krew list

To list all plugins install via `krew`, run:

    kubectl krew list

## Plugin Lifecycle

Plugins you are using might have newer versions available. To upgrade a single
plugin, run:

    kubectl krew upgrade <PLUGIN>

If you want to upgrade all plugins to their latest versions, run the same command
without any arguments:

    kubectl krew upgrade

Since `krew` itself is a plugin also managed through `krew`, running the upgrade
command may also upgrade your `krew` version.

**Note:** Plugins installed via `--HEAD` are always upgraded. This process
allows you to upgrade to the latest commit available in the source repository.


## Uninstalling Plugins

When you don't need a plugin anymore you can uninstall it with:

    kubectl krew remove <PLUGIN>


## Uninstalling Krew

Installing `krew` is as easy as deleting its installation directory.

To find `krew`'s installation directory, run:

    kubectl krew version

And delete the directory listed in `BasePath:` field. On macOS/Linux systems,
deleting the installation location can be done by executing:

    rm -rf ~/.krew
