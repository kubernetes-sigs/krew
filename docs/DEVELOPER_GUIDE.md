# Developer Guide

This guide is intended for plugin developers. If you are not developing kubectl
plugins, read the [User Guide](./USER_GUIDE.md) to learn how to use krew.

This guide explains how to package, test, run plugins locally and make them
available on the krew index.

<!-- TOC depthFrom:2 -->

- [Developing a Plugin](#developing-a-plugin)
    - [Packaging plugins for krew](#packaging-plugins-for-krew)
    - [Writing a plugin manifest](#writing-a-plugin-manifest)
        - [Specifying platform-specific instructions](#specifying-platform-specific-instructions)
        - [Specifying files to install](#specifying-files-to-install)
        - [Specifying plugin executable](#specifying-plugin-executable)
        - [Specifying a plugin download URL](#specifying-a-plugin-download-url)
- [Installing Plugins Locally](#installing-plugins-locally)
- [Publishing Plugins](#publishing-plugins)
    - [Submitting a plugin to krew](#submitting-a-plugin-to-krew)
    - [Updating existing plugins](#updating-existing-plugins)

<!-- /TOC -->

## Developing a Plugin

Before creating a plugin, read the [Kubernetes Plugins documentation][plugins].

**If you are writing a plugin in Go:** Consider using the
[cli-runtime](https://github.com/kubernetes/cli-runtime/) project which is
designed to provide the same command-line arguments, kubeconfig parser,
Kubernetes API REST client, and printing logic.

Below you will create a small plugin named `foo` which prints the environment
variables to the screen and exits.

Read the [Naming Guide](./NAMING_GUIDE.md) for choosing a name for your plugin.

Create an executable file named `kubectl-foo` with the following contents:

```bash
#!/usr/bin/env bash
env
```

Then make this script executable:

    chmod +x ./kubectl-foo

(Since this plugin requires `bash`, it will only work on Unix platforms. If you
need to support windows, develop a `kubectl-foo.exe` executable.)

Now, if you place this plugin to anywhere in your `$PATH` directories, you
should be able to call it like:

    kubectl foo

### Packaging plugins for krew

To package a plugin via krew, you need to:

1. Develop a plugin and archive it into a `.zip` or `.tar.gz`.
2. Write a [plugin manifest file](#writing-a-plugin-manifest).
3. Test installation locally with the archive and the manifest file.

To make a plugin available to everyone via krew, you need to:

1. Make the archive file (`.zip` or `.tar.gz`) **publicly downloadable** on a
   URL (you can host it yourself or use GitHub releases feature).
2. Submit the plugin manifest file to the [krew index][index] repository.

Plugin packages need to be available to download from the public Internet.
A service like
[GitHub Releases](https://help.github.com/articles/creating-releases/)
is recommended.

### Writing a plugin manifest

Each krew plugin has a "plugin manifest" file that lives in the [krew index
repository][index].

This file describes which files krew must copy out of your plugin archive and
how to run your plugin. It is also used to determine if an upgrade to your
plugin is available and how to install it.

For a plugin named `foo`, the plugin manifest file should be named `foo.yaml`.
An example manifest for this plugin that runs only on Linux and macOS looks
like this (see [krew index](https://github.com/kubernetes-sigs/krew-index/tree/master/plugins) for
all plugin manifests):

```yaml
apiVersion: krew.googlecontainertools.github.com/v1alpha2
kind: Plugin
metadata:
  name: foo               # plugin name must match your manifest file name (e.g. foo.yaml)
spec:
  version: "v0.0.1"       # optional, only for documentation purposes
  platforms:
  # specify installation script for linux and darwin (macOS)
  - selector:             # a regular Kubernetes selector
      matchExpressions:
      - {key: os, operator: In, values: [darwin, linux]}
    # url for downloading the package archive:
    uri: https://github.com/example/foo/releases/v1.0.zip
    # sha256sum of the file downloaded above:
    sha256: "208fde0b9f42ef71f79864b1ce594a70832c47dc5426e10ca73bf02e54d499d0"
    files:                     # copy the used files out of the zip archive
    - from: "/foo-*/unix/*.sh" # path to the files extracted from archive
      to: "."               # '.' refers to the root of plugin install directory
    bin: "./kubectl-foo"  # path to the plugin executable after copying files above
  shortDescription: Prints the environment variables.
  homepage: https://github.com/kubernetes-sigs/krew # optional, url for the project homepage
  # (optional) use caveats field to show post-installation recommendations
  caveats: |
    This plugin needs the following programs:
    * env(1)
  description: |
    This plugin shows all environment variables that get injected when
    launching a program as a plugin. You can use this field for longer
    description and example usages.
```

#### Specifying platform-specific instructions

krew makes it possible to install the same plugin on different operating systems
(like `windows`, `darwin` (macOS), and `linux`) and different architectures
(like `amd64`, `386`, `arm`).

To support multiple platforms, you may need to define multiple `platforms` in
the plugin manifest. The `selector` field matches to operating systems and
architectures using the keys `os` and `arch` respectively.

**Example:** Match to Linux:

```yaml
  platforms:
  - selector:
      matchLabels:
        os: linux
```

**Example:** Match to a Linux or macOS platform, any architecture:

```yaml
...
  platforms:
  - selector: # A regular Kubernetes label selector
      matchExpressions:
      - {key: os, operator: In, values: [darwin, linux]}
```

**Example:** Match to Windows 64-bit:

```yaml
  platforms:
  - selector:
      matchLabels:
        os: windows
        arch: amd64
```

The possible values for `os` and `arch`  come from the Go runtime. Run
`go tool dist list` to see all possible platforms and architectures.

#### Specifying files to install

Each operating system may require a different set of files from the archive to
be installed. You can use the `files` field in the plugin manifest to specify
which files should be copied into the plugin directory after extracting the
plugin archive.

The `file:` list specifies the copy operations (like `mv(1) <from> <to>`) to
the files `from` the archive `to` the installation destination.

**Example:** Copy all files in the `bin/` directory of the extracted archive to
the root of your plugin:

```yaml
    files:
    - from: bin/*.exe
      to: .
```

Given the file operation above, if the extracted plugin archive looked like
this:

```text
.
├── README.txt
├── README.txt
└── bin
│   └── kubectl-foo-linux
    └── kubectl-foo-windows.exe
```

The resulting installation directory would up just with:

```text
.
└── krew-foo-windows.exe
```

#### Specifying plugin executable

Each `platform` field requires a path to the plugin executable in the plugin's
installation directory.

krew creates a symbolic link to the plugin executable specified in the `bin`
field in `$HOME/.krew/bin/` (which all krew users need to add to their `$PATH`).

For example, if your plugin executable is named `start-foo.sh`, specify:

```yaml
platforms:
  - bin: "./start-foo.sh"
    ...
```

krew will create a [symbolic link](https://en.wikipedia.org/wiki/Symbolic_link)
named `kubectl-foo` (and `kubectl-foo.exe` on Windows) to your plugin executable
after installation is complete. The name of the symbolic link comes from the
plugin name.

> **Note on underscore conversion:** If your plugin name contains dashes, krew
> will automatically convert them to underscores for kubectl to be able to find
> your plugin.
>
> For example, if your  is named `view-logs` and your plugin binary is named
> `run.sh`, krew will create a symbolic named `kubectl-view_logs` automatically.

#### Specifying a plugin download URL

krew plugins must be packaged as `.zip` or `.tar.gz` archives and should be made
available for download publicly. 
Downloading from a URL also requires a checksum of the downloaded content:

- `uri`: URL to the archive file (`.zip` or `.tar.gz`)
- `sha256`: sha256 sum of the archive file

```yaml
  platforms:
  - uri: https://github.com/barbaz/foo/archive/v1.2.3.zip
    sha256: "29C9C411AF879AB85049344B81B8E8A9FBC1D657D493694E2783A2D0DB240775"
    ...
```

## Installing Plugins Locally

After you have:

- written your `<PLUGIN>.yaml`
- archived your plugin into a `.zip` or `.tar.gz` file

you can now test whether your plugin installs correctly on krew by running:

```bash
kubectl krew install --manifest=foo.yaml --archive=foo.tar.gz
```

- `--manifest` flag specifies a custom manifest rather than picking it up from
  the default [krew index][index]
- `--archive` overrides the download `url` specified in the plugin manifest and
  uses a local file instead.

If the installation fails, run the command with `-v=4` flag for verbose logs.
If your installation has succeeded, you should be able to run:

    kubectl foo

If you made your archive file available to download on the Internet, run the
same command without `--archive` and actually test downloading the file from
`url`.

If you need other `platforms` definitions that don't match your current machine,
you can use `KREW_OS` and/or `KREW_ARCH` environment variables. For example,
if you're on a Linux machine, you can test Windows installation with:

    KREW_OS=windows krew install --manifest=[...]

After you have tested your plugin, uninstall it with `kubectl krew uninstall foo`.

## Publishing Plugins

### Submitting a plugin to krew

After you have tested that the plugin can be installed and works you should
create a pull request to the [Krew Index][index] with your `<PLUGIN>.yaml`
manifest file.

The new plugin file should be submitted to the `plugins/` directory in the
index repository.

After the pull request gets accepted into the main index, the plugin will be
available for all users.

Please make sure to include dependencies of your plugin and extra configuration
needed to run the plugin in the `caveats:` field.

### Updating existing plugins

When you have a newer version of your plugin, create a new pull request that
updates `uri` and `sha256` fields of the plugin manifest file.

Optionally, you can use the `version` field to match to your plugin's released
version string.

[index]: https://github.com/kubernetes-sigs/krew-index
[plugins]: https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/
