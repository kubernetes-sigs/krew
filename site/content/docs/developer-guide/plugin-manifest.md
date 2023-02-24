---
title: Writing Krew plugin manifests
slug: plugin-manifest
weight: 150
---

{{<toc>}}

Each Krew plugin has a "plugin manifest", which is a YAML file that describes
the plugin, how it can be downloaded, and how it is installed on a machine.

Plugin manifests must have the same name as the plugin (e.g. plugin `foo` has a
manifest file named `foo.yaml`).

It is **highly recommended** you look at [example plugin manifests]({{< ref
"example-plugin-manifests.md" >}}), copy an existing plugin manifest, and adapt it
to your needs rather than starting from scratch.

## Sample plugin manifest

Here's a sample manifest file for a bash-based plugin that supports
only Linux and macOS:

```yaml
apiVersion: krew.googlecontainertools.github.com/v1alpha2
kind: Plugin
metadata:
  # 'name' must match the filename of the manifest. The name defines how
  # the plugin is invoked, for example: `kubectl restart`
  name: restart
spec:
  # 'version' is a valid semantic version string (see semver.org). Note the prefix 'v' is required.
  version: "v0.0.1"
  # 'homepage' usually links to the GitHub repository of the plugin
  homepage: https://github.com/achanda/kubectl-restart
  # 'shortDescription' explains what the plugin does in only a few words
  shortDescription: "Restarts a pod with the given name"
  description: |
    Restarts a pod with the given name. The existing pod
    will be deleted and created again, not a true restart.
  # 'platforms' specify installation methods for various platforms (os/arch)
  # See all supported platforms below.
  platforms:
  - selector:
      matchExpressions:
      - key: "os"
        operator: "In"
        values:
        - darwin
        - linux
    # 'uri' specifies .zip or .tar.gz archive URL of a plugin
    uri: https://github.com/achanda/kubectl-restart/archive/v0.0.3.zip
    # 'sha256' is the sha256sum of the url (archive file) above
    sha256: d7079b79bf4e10e55ded435a2e862efe310e019b6c306a8ff04191238ef4b2b4
    # 'files' lists which files should be extracted out from downloaded archive
    files:
    - from: "kubectl-restart-*/restart.sh"
      to: "./restart.sh"
    - from: "LICENSE"
      to: "."
    # 'bin' specifies the path to the the plugin executable among extracted files
    bin: restart.sh
```

## Documentation fields

- `shortDescription:` (required): Tagline for your plugin. It should not exceed
  a few words (to be properly shown in `kubectl krew search` output). **Avoid**
  using redundant words.

- `description:` (required): Explain what your plugin does at a high level to help users
  decide if they want to install the plugin. **Avoid** including a list of
  commands or options for your plugin; implement `-h`/`--help` in your plugin
  instead.

- `caveats:` Use this field if your plugin needs an extra step to work; for example, if it
  needs certain programs to be installed on a user’s system, list them here.
  **Avoid** using this field as a documentation field.

  `caveats` are shown to the user after installing the plugin for the first time.

## Specifying plugin download options

Krew plugins must be packaged as `.zip` or `.tar.gz` archives, and should be
accessible to download from a user’s machine. The relevant fields are:

- `uri`: URL to the archive file (`.zip` or `.tar.gz`)
- `sha256`: sha256 sum of the archive file

```yaml
  platforms:
  - uri: https://github.com/foo/bar/archive/v1.2.3.zip
    sha256: "29C9C411AF879AB85049344B81B8E8A9FBC1D657D493694E2783A2D0DB240775"
    ...
```

## Specifying platform-specific instructions

Krew makes it possible to install the same plugin on different operating systems
(e.g., `windows`, `darwin` (macOS), and `linux`) and different architectures
(e.g., `amd64`, `386`, `arm`, `arm64` and `ppc64le`).

**All supported platforms:**

The supported platforms for plugins are the ones that Krew itself is distributed in.
See all supported platforms on the [releases page](https://github.com/kubernetes-sigs/krew/releases).

To support multiple platforms, you may need to define multiple `platforms` in
the plugin manifest. The `selector` field matches to operating systems and
architectures using the keys `os` and `arch` respectively.

**Example:** Match to a Linux or macOS platform, any architecture:

```yaml
...
  platforms:
  - selector:
      matchExpressions:
      - {key: "os", operator: "In", values: [darwin, linux]}
```

**Example:** Match to Windows, 64-bit:

```yaml
  platforms:
  - selector:
      matchLabels:
        os: windows
        arch: amd64
```

**Example:** Match to Linux (any architecture):

```yaml
  platforms:
  - selector:
      matchLabels:
        os: linux
```

The possible values for `os` and `arch` come from the Go runtime. Run
`go tool dist list` to see all possible platforms and architectures.

## Specifying files to install

Each operating system may require a different set of files from the archive to
be installed.

You can use the `files` field to specify
the files that should be copied into the plugin directory while extracting the
files from the archive.

* **Example:** Extract all files. If the `files:` list is unspecified, it defaults to:

  ```yaml
  files:
  - from: *
    to: .
  ```

  > **Note:** If applicable, it is a best practice to skip the `files:` list altogether.

* **Example:** Copy specific files.

  ```yaml
  files:
  - from: LICENSE
    to: .
  - from: bin/foo-windows.exe
    to: foo.exe
  ```

* **Example:** Use wildcards to match multiple files.

  ```yaml
  files:
  - from: "*/*.sh"
    to: "."
  ```

  This operation preserves the files' directory structure in the destination directory.

## Specifying the plugin executable

Each `platform` field requires a path to the plugin executable in the plugin's
installation directory.

For example, if your plugin executable is named `foo.sh`, specify:

```yaml
platforms:
  - bin: "./foo.sh"
    ...
```

Krew creates a [symbolic link](https://en.wikipedia.org/wiki/Symbolic_link)
named `kubectl-foo` (and `kubectl-foo.exe` on Windows) to your plugin executable
after installation is complete. The name of the symbolic link comes from the
plugin name (specifically the `metadata.name` field).

> **Note on underscore conversion:** If your plugin name contains dashes, Krew
> will automatically convert them to underscores to enable
> [kubectl](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/#names-with-dashes-and-underscores)
> to find your plugin.
>
> For example, if your plugin name is `view-logs` and your plugin binary is named
> `run.sh`, Krew will create a symbolic link named `kubectl-view_logs` automatically.

> **Note on Windows entrypoints:** Currently only `.exe` entrypoints are supported
> on Windows, `.bat` and `.ps1` are not supported, refer to [this issue](https://github.com/kubernetes-sigs/krew/issues/131).
