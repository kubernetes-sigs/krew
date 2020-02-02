# User Guide

This guide shows how to use `krew` as a user after installing it.

<!-- TOC depthFrom:2 -->

- [Discovering Plugins](#discovering-plugins)
- [Installing Plugins](#installing-plugins)
- [Listing Installed Plugins](#listing-installed-plugins)
- [Upgrading Plugins](#upgrading-plugins)
- [Uninstalling Plugins](#uninstalling-plugins)
- [Uninstalling Krew](#uninstalling-krew)

<!-- /TOC -->

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
URI: https://github.com/ahmetb/kubectl-extras/archive/c403c57.zip
SHA256: 8be8ed348d02285abc46bbf7a4cc83da0ee9d54dc2c5bf86a7b64947811b843c
DESCRIPTION:
 Pretty print the current cluster certificate.
 The plugin formats the certificate in PEM following RFC1421.
VERSION: v1.0.0
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

```sh
kubectl ca-cert
```

## Listing Installed Plugins

All plugins available to `kubectl` (including those not installed via `krew`) can
be listed using:

    kubectl plugin list

To list all plugins installed via `krew`, run:

    kubectl krew list

## Upgrading Plugins

Plugins you are using might have newer versions available. To upgrade a single
plugin, run:

    kubectl krew upgrade <PLUGIN>

If you want to upgrade all plugins to their latest versions, run the same command
without any arguments:

    kubectl krew upgrade

Since `krew` itself is a plugin also managed through `krew`, running the upgrade
command may also upgrade your `krew` version.

### Krew upgrade check

When using krew, it will check if a new version of krew is available once a day
by calling the GitHub API. If you want to opt out of this feature, you can set
the `KREW_NO_UPGRADE_CHECK` environment variable. To permanently disable this,
add the following to your `~/.bashrc`, `~/.bash_profile`, or `~/.zshrc`:

    export KREW_NO_UPGRADE_CHECK=1

## Uninstalling Plugins

When you don't need a plugin anymore you can uninstall it with:

    kubectl krew uninstall <PLUGIN>

## Uninstalling Krew

Uninstalling `krew` is as easy as deleting its installation directory.

To find `krew`'s installation directory, run:

    kubectl krew version

And delete the directory listed in `BasePath:` field. On macOS/Linux systems,
deleting the installation location can be done by executing:

    rm -rf ~/.krew
