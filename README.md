# krew

krew is the missing kubectl plugin manager.

## What is krew?

krew is a tool that makes it easy to install
[kubectl plugins](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/). 
krew helps you discover plugins, install and manage them on your machine. It is
similar to tools like apt, dnf or [brew](http://brew.sh).

### Installation

For macOS and Linux:

- Make sure that git is installed.
- Paste this command to your terminal:

```bash
(
 set -x; cd "$(mktemp -d)" &&
 curl -fsSLO "https://storage.googleapis.com/krew-test/krew.zip" &&
 unzip krew.zip &&
 "./build/krew-$(uname | tr '[:upper:]' '[:lower:]')" install krew
)
```

Windows:

1. Make sure that git is installed
2. Download https://storage.googleapis.com/krew-test/krew.zip
3. Unzip the file
4. Launch a command-line window in the extracted directory
5. Run: ./krew-windows.exe install krew

To verify the installation run `kubectl plugin`.
You should see new subcommands.
Run `kubectl plugin list` to see all installed plugins.

### Finding plugins

This command shows all the plugins available in krew index:

```bash
kubectl plugin search
```

### Installing plugins

Choose one of the plugins from the list returned in the previous command,
for example:

```bash
kubectl plugin install ca-cert
```

This plugin ("ca-cert") prints the CA cert of the current cluster as PEM.
Execute this plugin by running the command:

```bash
kubectl plugin ca-cert
```

### Uninstalling a plugin

```bash
kubectl plugin remove ca-cert
```

### Documentation

Read the complete [User Guide](./docs/USER_GUIDE.md) for more details.

## Publishing Plugins

To publish your plugin on krew, you need to make the releases available for
download, and contribute a plugin descriptor file to krew-index repository.

Read the [Plugin Developer Guide](./docs/DEVELOPER_GUIDE.md) for details.

# Additional Links

- [Architecture](./docs/KREW_ARCHITECTURE.md)
- [Docs](./docs/)
- [Contributing](./CONTRIBUTING.md)  

# LICENSE

The code is submitted under the Apache 2.0 License described in the
[LICENSE](./LICENSE) file.

----

This is not an official Google project.
