# User Guide

This guide will help you with your first steps using krew.

## Discover Plugins

To find plugins run the `kubectl plugin search` command.
This command lists all available plugins:

```text
$ kubectl plugin search
NAME               SHORT                                              STATUS
ca-cert            Print PEM CA certificate of current cluster        available
extract-context    Extract current-context on kubectl as a kubecon... available
krew               Install plugins                                    available
mtail              Tail logs from multiple pods matching label sel... available
view-secret        Decode secrets                                     available
...
```

Specifying keywords searches in the plugin list:

```text
$ kubectl plugin search crt
NAME               SHORT                                       STATUS
ca-cert            Print PEM CA certificate of current cluster available
view-secret        Decode secrets                              available
```

To get more information on the "ca-cert" plugin,
run `kubectl plugin info ca-cert`.

```text
$ kubectl plugin info ca-cert
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

## Install Plugins

Now, install the plugin by running `kubectl plugin install` command.
The installation downloads the plugin from the versioned URL and verify its
contents with sha256. As it is shown the ca-cert plugin also specifies a HEAD,
which should always point to the latest location. Uri is always favoured if
HEAD also exists. If just wither exist it is be the default.
You want to install the HEAD, to do so force it with `--HEAD=true`.

```text
$ kubectl plugin install ca-cert --HEAD=true
Will install plugin: ca-cert
Installed plugin: ca-cert
CAVEATS:
 This plugin needs the following programs:
 * base64
```

The "CAVEATS" section in the output describes noteworthy details about running
the plugin, such as dependencies that you need to install.
Our plugin just needs the base64 tool.

To show all installed plugins you can run,

```text
$ kubectl plugin list
PLUGIN  VERSION
ca-cert HEAD
krew    6e8a790d8a34885cd40436761f564d94b74d46b314eed5bd02054654946034ef
```

As you can see there are two plugins installed, `krew` itself and `ca-cert`.

Now the plugin can be tried out with,

```text
$ kubectl plugin ca-cert
-----BEGIN CERTIFICATE-----
...
```

You should also see it as a subcommand of `kubectl plugin`.

## Plugin Lifecycle

Plugins you are using might have newer versions available.
You can run `kubectl plugin upgrade ca-cert` to only upgrade the `ca-cert` plugin.

This time you want to upgrade everything, including krew so just run
`kubectl plugin upgrade`.

```text
$ kubectl plugin upgrade
Upgraded plugin: ca-cert
Skipping plugin krew, it is already on the newest version
```

As you can see krew upgraded krew and `ca-cert`, plugins that are already on the
latest version are skipped. HEAD installations are always be upgraded and
stay HEAD. This process allows you to always have the newest plugins and
keep krew up to date.

Krew itself is a plugin which is also managed through `krew`.
This allows krew to not rely on other package managers.
Krew controls it's own lifecycle.

## Remove Plugins

When you don't need a plugin anymore you can uninstall it with 
`kubectl plugin remove`.

```text
$ kubectl plugin remove ca-cert
Removed plugin ca-cert
```

## Uninstalling Krew

Run command `kubectl plugin krew version`
and remove the `BasePath` directory listed in the output to uninstall `krew`
and all plugins that were installed with it.