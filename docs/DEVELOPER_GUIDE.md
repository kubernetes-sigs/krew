# Developer Guide

This guide is intended for plugin developers. If you are not developing kubectl
plugins, read the [User Guide](./USER_GUIDE.md).

This guide explains how to package, test, run plugins locally and make them
available on the krew index.

## Creating a Plugin

Please read the
[Official plugin docs](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)
to learn about plugins.

This document shows how to create a plugin named `foo`.
The plugin shows the environment variables on unix and windows systems.

First, create a public GitHub repository for the plugin.
In this repository there should be two plugin files:

`unix/plugin.yaml`

```yaml
name: foo
shortDesc: Example plugin
command: env
```

`windows/plugin.yaml`:

```yaml
name: foo
shortDesc: Example plugin
command: SET
```

Each plugin needs a plugin descriptor.
Those files are called plugin descriptors, they describe the entrypoint of the 
kubectl plugin. 

Commit and push the two new files to your public repository. 

### Making Plugins Publicly Accessible

Plugin packages need to be available to download from the public Internet.
A service like
[GitHub Releases](https://help.github.com/articles/creating-releases/)
is recommended.
It is also possible to get the latest release for a GitHub repository from the
URL: `https://github.com/<user>/<project>/archive/master.zip`.

### Writing a Plugin Index File

Each krew plugin has a "plugin index manifest" file that lives in the index
repository. This file describes which files to install from the provided archive.
This file is different than the `plugin.yaml` file.
The index plugin index manifest allows krew to manage your plugin
installation this file. 

### Example Plugin Index File

```yaml
apiVersion: krew.googlecontainertools.github.com/v1alpha1
kind: Plugin
metadata:
  name: foo
spec:
  platforms:
  - selector: # A regular Kubernetes label selector
      matchExpressions:
      - {key: os, operator: In, values: [macos, linux]} 
    head: https://github.com/barbaz/foo/archive/master.zip
    # This is used during installation. It uses file Globs to copy required files.
    files:
    - from: "/unix/*"
      to: "."
  - selector:
      matchLabels:
        os: windows
    head: head: https://github.com/barbaz/foo/archive/master.zip
    files: 
    - from: "/windows/*"
      to: "."
  # Version does not follow any conventions and is not functional.
  version: "v0.0.1"
  shortDescription: Short description of foo
  description: |
      This plugin shows all environment
      variables that get injected when
      launching a program as a plugin.
      All environment variables are
      prefixed with KUBECTL_*
  caveats: |
    This plugin needs the following programs:
    * env (unix)
    * SET (windows)
```

Choose a name for your plugin.
A plugin name must be unique within the krew index.
The name can contain alphanumeric characters and dashes.

The following YAML file, named `foo.yaml`,
shows the manifest for a plugin named `foo`:

```yaml
apiVersion: krew.googlecontainertools.github.com/v1alpha1
kind: Plugin
metadata:
  name: foo
...
```

---

To give some more information about a plugin provide a
`shortDescription` and `description`:

```yaml
...
  shortDescription: Short description of foo
  description: |
      This plugin shows all environment
      variables that get injected when
      launching a programm as a plugin.
      All environment variables are
      prefixed with KUBECTL_*
...
```

---

The version field is not used by krew and is not required; however,
you can use this field to provide users a way to show the version number of
your plugin.

```yaml
...
  version: "v0.0.1"
...
```

---

To allow a plugin to work on different platforms, you can specify different
target platforms those are stored in the `platforms` array:

```yaml
...
  platforms:
  - selector: # A regular Kubernetes label selector
      matchExpressions:
      - {key: os, operator: In, values: [macos, linux]} 
    ...
  - selector:
      matchLabels:
        os: windows
    ...
...
```

This plugin works on *Linux*, *macOS* and *Windows*.
Krew uses Kubernetes
[Label Selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
to match for the platform specific keys `os` and `arch`.
The values come from golang's
[GOOS and GOARCH](https://golang.org/pkg/runtime/#pkg-constants).
The label selectors are evaluated on the user's machine during the installation.

---

Each operating system may require a different set of files from the
archive to be installed. You can use the `files` field to specify them.
The `file:` list specifies operations (like `mv(1) <from> <to>`) to copy
the files `from` the archive `to` the installation destination.

After executing the move operations,
the resulting directory must contain a plugin descriptor file (`plugin.yaml`).

```yaml
...
    files:
    - from: "/unix/*"
      to: "."
...
```

This file operation moves all files from the `/unix/*` directory to the
root of the installation directory.

Given the file operation above, assume the plugin archive looks like this:

```text
.
├── unix
│   └── plugin.yaml
└── windows
    └── plugin.yaml
```

The resulting installation directory will look like:

```text
.
└── plugin.yaml
```

---

There are two ways to specify a plugin archive location:

1. Download from a URL pointing and verify its checksum with sha256.
   This uses `url` and `sha256` fields.
2. Download from a URL without verifying its checksum.
   This uses the `head` field. You can use this for development.

Versioned files have two fields that need to be specified:

1. The `sha256` hash of the archive that is will be downloaded.
2. The `uri` where the archive can be found. 

When you specify a HEAD, it is enough to enter the `head` url.
it is intended that the HEAD points archive always to the latest release.
The checksum of the archive file specified on `head` won't be verified. 

```yaml
...
    head: https://github.com/barbaz/foo/archive/master.zip
    uri: https://github.com/barbaz/foo/archive/v1.2.3.zip
    sha256: "29C9C411AF879AB85049344B81B8E8A9FBC1D657D493694E2783A2D0DB240775"
...
```

### Running the Plugin

To test the plugin locally, you can install the plugin with:

```bash
kubectl plugin install -v=4 --source=./foo.yaml
```

This will install the `foo` plugin.
To see the plugin directory, get the `InstallPath`:

```bash
kubectl plugin krew version
```

The installation target directory for the `foo` plugin is
`<InstallPath>/<PluginName>/<PluginVersion>/`.
There should always be only one version directory.

You can now run your plugin!

```bash
kubectl plugin foo
```

### Cleaning up

After you have tested the plugin, remove it with `kubectl plugin remove foo`.

### Publishing the Plugin

After you have tested that the plugin can be installed and works you should
create a pull request for the [Main Index](https://github.com/GoogleContainerTools/krew-index).
After the pull request gets accepted into the main index, the plugin will be available for
all users.

Please make sure to include dependencies and extra configuration to run the
plugin in the `caveats` section.
The new plugin file should be submitted to the `plugins/` directory in the index.

### Updating a Published Plugin

Create a pull request with the updated `uri` and `sha256`,
it is also useful to change the `version` field so that users can distinguish
the different versions.


