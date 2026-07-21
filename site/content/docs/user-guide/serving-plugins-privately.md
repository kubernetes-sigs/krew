---
title: Serving Plugins Privately
slug: private-plugins
weight: 850
---

Plugin archives are the binary artifacts that contain the compiled plugin executable.  
These archives are typically hosted in a public registry (such as the official krew-index).  

For plugins stored in **private registries** that require authentication, Krew supports
reading credentials from the user's `.netrc` file (`_netrc` on Windows).  

To enable this behavior, use the `--enable-netrc` flag.

This is commonly required when using [custom indexes]({{<ref "using-custom-indexes.md">}}),
where plugin artifacts are served from a private registry that requires authentication.

For example, to install a plugin named `bar` from custom index `foo`:

```sh
{{<prompt>}}kubectl krew install --enable-netrc foo/bar
```

The host portion of the plugin archive URL (as specified in the custom index)
must have a corresponding entry in your `.netrc` file with the appropriate credentials:

```text
machine <private registry host>
  login <username>
  password <password>
```

By default, Krew looks for the `.netrc` file in your home directory:

- Linux & macOS: `~/.netrc`
- Windows: `%HOME%\_netrc`

You can override the default location by using the `--netrc-file` flag.

## Serving plugins from an OCI registry

A plugin archive can also be stored in an OCI registry (GitHub Container
Registry, Docker Hub, Harbor, a cloud provider's registry...) as an artifact
whose single layer is the archive file itself. The plugin manifest points at
it with an `oci://` URI:

```yaml
  platforms:
  - uri: oci://ghcr.io/<org>/plugins/kubectl-foo:v0.1.0_linux_amd64
    sha256: <sha256 of the archive file>
    ...
```

Push the archive as it is, for example with [ORAS](https://oras.land):

```sh
{{<prompt>}}oras push ghcr.io/<org>/plugins/kubectl-foo:v0.1.0_linux_amd64 kubectl-foo_linux_amd64.tar.gz
```

The layer holds the archive's exact bytes, so the manifest's `sha256` is the
same one a `https://` download would verify. The artifact must have exactly one
layer.

Krew authenticates to the registry with your Docker credentials
(`~/.docker/config.json`, credential helpers included): if `docker login` works
for the registry, `kubectl krew install` works too. No flag is needed, and who
can install is managed with the registry's own access control.

## Serving plugins from a Private GitHub repository

Below is a reference on how Krew artifacts can be stored in a private GitHub repository, and
retrieved using a GitHUb PAT (personal access token) stored in a `.netrc` file.

Example repository structure:

```text
internal-krew/
├── artifacts/                            ← compiled plugin artifacts (stored in branches, or in main)
│   └── <plugin-name>/<version>/          ← e.g. artifacts/<plugin-name>/v0.1.0/kubectl-<plugin-name>-v0.1.0.tar.gz
│       └── kubectl-<plugin-name>-v0.1.0.tar.gz
├── plugins/                              ← Krew plugin manifests
│   └── <plugin-name>.yaml                ← manifest for kubectl-<plugin-name>
├── src/                                  ← source code for all plugins
│   ├── <plugin-name>/                    ← kubectl-<plugin-name> source code
│   │   └── .krew.yaml                    ← Krew manifest template
│   └── <other-plugin>/                   ← additional plugins follow the same pattern
├── .gitignore
└── README.md
```

A release pipeline can be used to build the plugin artifact and store it under `artifacts/`, the
plugin manifest is stored under `plugins/`.

### Key patterns

Plugins code live under `src/<plugin-name>/`

Krew template (`.krew.yaml`) used to populate the plugin manifest

Krew Manifests are stored `plugins/<plugin-name>.yaml` (generated from the `.krew.yaml` template)

Artifacts are published to the following locations in the repository

- `artifacts/<plugin-name>/<version>/` in the `main` branch, or 
- `artifacts/<plugin-name>/<version>/` in a dedicated branch named `artifacts-<plugin-name>-<version>`

This structure allows direct downloads via `raw.githubusercontent.com` URIs — even from private repositories — using GitHub Personal Access Token (PAT) authentication.

e.g. `artifacts-foo-v0.1.0` plugin archive URI:

```text
https://raw.githubusercontent.com/<org>/internal-krew/<branch>/artifacts/foo/v0.1.0/kubectl-foo-v0.1.0.tar.gz
```

Note: GitHub release artifacts on private repos do not support this auth method, that's why
`raw.githubusercontent.com` is used.

### Authentication

Users need to generate a GitHub fine-grained (repository scoped) PAT (Personal Access Token) with
read-only permissions: `contents` & `metadata`

The produced PAT token is stored in the `.netrc` file:

```text
machine github.com
  login token
  password  <fine-grained github PAT>

machine raw.githubusercontent.com
  login token
  password <fine-grained github PAT>
```
