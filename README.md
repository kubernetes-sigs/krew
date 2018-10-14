# krew

krew is the kubectl plugin manager.

## What is krew?

krew is a tool that makes it easy to install
[kubectl plugins](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).
krew helps you discover plugins, install and manage them on your machine. It is
similar to tools like apt, dnf or [brew](http://brew.sh).

- **For kubectl users:** krew helps you find, install and manage kubectl plugins
  in a consistent way.
- **For plugin developers:** krew helps you package and distribute your plugins
  on multiple platforms and makes them discoverable.

krew is easy to use:

```sh
kubectl krew search               # show all plugins
kubectl krew install view-secret  # install a plugin named "view-secret"
kubectl view-secret               # use the plugin
kubectl upgrade                   # upgrade installed plugins
kubectl remove view-secret        # uninstall a plugin
```

Read the [User Guide](./docs/USER_GUIDE.md) for detailed documentation.

### Installation

> :warning: **Warning** :warning: `krew` is not yet compatible with kubectl 1.12
> ([#33](https://github.com/GoogleContainerTools/krew/issues/33)). If you want
> to try out krew, use [v0.1](https://github.com/GoogleContainerTools/krew/tree/v0.1.1#installation) with
> kubectl v1.11.x or older.

For macOS and Linux:

1. Make sure that `git` is installed.
2. Run this command in your terminal to download and install `krew`:

    ```sh
    (
      set -x; cd "$(mktemp -d)" &&
      curl -fsSLO "https://github.com/GoogleContainerTools/krew/releases/download/v0.2.0/krew.{tar.gz,yaml}" &&
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

Windows:

1. Make sure `git` is installed on your system.
1. Download `krew.zip` and `krew.yaml` from the [Releases][releases] page.
1. Extract the `krew.zip` archive to a directory, navigate to the directory.
1. Launch a command-line window (`cmd.exe`) in that directory.
1. Run the following command to install krew (pass the correct
   paths to `krew.yaml` and `krew.zip` below):

       .\krew-windows_amd64.exe install --source=krew.yaml --archive=krew.zip

3. Add `%USERPROFILE%\.krew\bin` to your `PATH` environment variable
   ([how?](https://java.com/en/download/help/path.xml))

[releases]: https://github.com/GoogleContainerTools/krew/releases

### Verifying installation

Run `kubectl plugin list` command to see installed plugins. This command should show `kubectl-krew` in the results. You can now use `kubectl krew` command.

### Documentation

Read the complete [User Guide](./docs/USER_GUIDE.md) for more details.

- [Documentation](./docs/)
- [Architecture](./docs/KREW_ARCHITECTURE.md)
- [Contributing](./CONTRIBUTING.md)

## Publishing Plugins

As a kubectl plugin developer, you need to:

1. make your plugin archive (.zip or .tar.gz) available to download
2. write a plugin manifest (.yaml) file and submit it to the [krew-index][index]

Read the [Plugin Developer Guide](./docs/DEVELOPER_GUIDE.md) for details.

# Roadmap

- **Support Multiple Index Repositories:** Tracked under
  [#23](https://github.com/GoogleContainerTools/krew/issues/23)
- **Donating krew to the SIG-CLI:** We plan to donate krew to the
  [SIG-CLI](https://github.com/kubernetes/community/tree/master/sig-cli). We
  have created a [KEP](https://github.com/kubernetes/community/pull/2340) that
  covers our intentions. Accepting the KEP means that kubectl will implement
  krew commands natively, and support the plugin format.

# LICENSE

The code is submitted under the Apache 2.0 License described in the
[LICENSE](./LICENSE) file.

----

This is not an official Google project.
