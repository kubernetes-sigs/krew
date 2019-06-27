# krew

[![Build Status](https://travis-ci.org/kubernetes-sigs/krew.svg?branch=master)](https://travis-ci.org/kubernetes-sigs/krew)
[![Code Coverage](https://codecov.io/gh/kubernetes-sigs/krew/branch/master/graph/badge.svg)](https://codecov.io/gh/kubernetes-sigs/krew)
[![Go Report Card](https://goreportcard.com/badge/kubernetes-sigs/krew)](https://goreportcard.com/report/kubernetes-sigs/krew)
[![LICENSE](https://img.shields.io/github/license/kubernetes-sigs/krew.svg)](https://github.com/kubernetes-sigs/krew/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/release-pre/kubernetes-sigs/krew.svg)](https://github.com/kubernetes-sigs/krew/releases)

krew is the package manager for kubectl plugins.

## What is krew?

krew is a tool that makes it easy to use [kubectl
plugins](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/). krew
helps you discover plugins, install and manage them on your machine. It is
similar to tools like apt, dnf or [brew](http://brew.sh).

- **For kubectl users:** krew helps you find, install and manage kubectl plugins
  in a consistent way.
- **For plugin developers:** krew helps you package and distribute your plugins
  on multiple platforms and makes them discoverable.

krew is easy to use:

```sh
kubectl krew search                 # show all plugins
kubectl krew install view-secret    # install a plugin named "view-secret"
kubectl view-secret                 # use the plugin
kubectl krew upgrade                # upgrade installed plugins
kubectl krew uninstall view-secret  # uninstall a plugin
```

Read the [User Guide](./docs/USER_GUIDE.md) for detailed documentation.

### Installation

> :warning: **Warning:** krew is only compatible with kubectl v1.12 or higher.

**macOS and Linux:**

#### Bash and ZSH

1. Make sure that `git` is installed.
2. Run this command in your terminal to download and install `krew`:

    ```sh
    (
      set -x; cd "$(mktemp -d)" &&
      curl -fsSLO "https://storage.googleapis.com/krew/v0.2.1/krew.{tar.gz,yaml}" &&
      tar zxvf krew.tar.gz &&
      ./krew-"$(uname | tr '[:upper:]' '[:lower:]')_amd64" install \
        --manifest=krew.yaml --archive=krew.tar.gz
    )
    ```
3. Add `$HOME/.krew/bin` directory to your PATH environment variable. To do
   this, update your `.bashrc` or `.zshrc` file and append the following line:

     ```sh
     export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"
     ```

   and restart your shell.

#### Fish

1. Make sure that `git` is installed.
2. Run this command in your terminal to download and install `krew`:

    ```fish
    begin
      set -x; set temp_dir (mktemp -d); cd "$temp_dir" &&
      curl -fsSLO "https://storage.googleapis.com/krew/v0.2.1/krew.{tar.gz,yaml}" &&
      tar zxvf krew.tar.gz &&
      set KREWNAME krew-(uname | tr '[:upper:]' '[:lower:]')_amd64 &&
      ./$KREWNAME install \
        --manifest=krew.yaml --archive=krew.tar.gz &&
      set -e KREWNAME; set -e temp_dir
    end
    ```
3. Add `$HOME/.krew/bin` directory to your PATH environment variable. To do
   this, update your `config.fish` file and append the following line:

     ```fish
     set -gx PATH $PATH $HOME/.krew/bin
     ```

   and restart your shell.


**Windows:**

1. Make sure `git` is installed on your system.
1. Download `krew.zip` and `krew.yaml` from the [Releases][releases] page.
1. Extract the `krew.zip` archive to a directory, navigate to the directory.
1. Launch a command-line window (`cmd.exe`) in that directory.
1. Run the following command to install krew (pass the correct
   paths to `krew.yaml` and `krew.zip` below):

       .\krew-windows_amd64.exe install --manifest=krew.yaml --archive=krew.zip

1. Add `%USERPROFILE%\.krew\bin` to your `PATH` environment variable
   ([how?](https://java.com/en/download/help/path.xml))

[releases]: https://github.com/kubernetes-sigs/krew/releases

### Verifying installation

Run `kubectl plugin list` command to see installed plugins. This command should
show `kubectl-krew` in the results. You can now use `kubectl krew` command.

### Upgrading krew

Since krew itself is installed as a "kubectl plugin" managed by krew, it can be
upgraded like a plugin by running the `kubectl krew upgrade` command.

## Documentation

- **Users:**
  - Read the [**User Guide**](./docs/USER_GUIDE.md) to learn how to use krew.
- **Plugin Developers:**
  - [Developer Guide](./docs/DEVELOPER_GUIDE.md): how to package and publish a
    plugin for krew.
  - [Naming Guide](./docs/NAMING_GUIDE.md): how to choose a good name for your
    plugin
- **Krew Developers:**
  - Building Krew (not written yet)
  - [Releasing Krew](./docs/RELEASING_KREW.md): how to release new version of
    krew.
  - [Plugin Lifecycle](./docs/PLUGIN_LIFECYCLE.md): how krew installs/upgrades
    plugins and itself.
  - [Krew Architecture](./docs/KREW_ARCHITECTURE.md): architectural decisions
    behind designing krew.

Visit [`./docs`](./docs) for all documentation.

## Roadmap

Please check out the [Issue Tracker](./issues) to see the plan of record for
new features and changes.

## Community

### Bug reports

* If you have a problem with the Krew itself, please file an
  issue in this repository.
* If you're having a problem with a particular plugin's installation or
  upgrades, file an issue at [krew-index][index] repository.
* If you're having an issue with an installed plugin, file an issue for the
  repository the plugin's source code is hosted at.

### Communication channels

* [Slack](https://kubernetes.slack.com/messages/sig-cli) #sig-cli
* [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-sig-cli)
* [Kubernetes Community site](http://kubernetes.io/community/)

### Contributing

Interested in contributing to Krew? Please refer to our 
[Contributing Guidelines](./docs/CONTRIBUTOR_GUIDE.md) for more details

### Code of Conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code
of Conduct](https://github.com/kubernetes-sigs/kustomize/blob/master/code-of-conduct.md).

[index]:https://github.com/kubernetes-sigs/krew-index
