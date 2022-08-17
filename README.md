<img src="assets/logo/horizontal/color/krew-horizontal-color.png" width="480"
  alt="Krew logo"/>

# Krew

[![Build Status](https://github.com/kubernetes-sigs/krew/workflows/Kubernetes-sigs/krew%20CI/badge.svg)](https://github.com/kubernetes-sigs/krew/actions)
[![Go Report Card](https://goreportcard.com/badge/kubernetes-sigs/krew)](https://goreportcard.com/report/kubernetes-sigs/krew)
[![LICENSE](https://img.shields.io/github/license/kubernetes-sigs/krew.svg)](https://github.com/kubernetes-sigs/krew/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/release-pre/kubernetes-sigs/krew.svg)](https://github.com/kubernetes-sigs/krew/releases)
![GitHub stars](https://img.shields.io/github/stars/kubernetes-sigs/krew.svg?label=github%20stars&logo=github)

Krew is the package manager for kubectl plugins.

## What does Krew do?

Krew is a tool that makes it easy to use [kubectl
plugins](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/). Krew
helps you discover plugins, install and manage them on your machine. It is
similar to tools like apt, dnf or [brew](https://brew.sh). Today, over [200
kubectl plugins][list] are available on Krew.

- **For kubectl users:** Krew helps you find, install and manage kubectl plugins
  in a consistent way.
- **For plugin developers:** Krew helps you package and distribute your plugins
  on multiple platforms and makes them discoverable.

## [Documentation][website]

Visit the [**Krew documentation**][website] to find **Installation**
instructions, **User Guide** and **Developer Guide**.

You can follow the [**Quickstart**][quickstart] to get started with Krew.

[website]: https://krew.sigs.k8s.io/
[quickstart]: https://krew.sigs.k8s.io/docs/user-guide/quickstart/

## Contributor Documentation

- [Releasing Krew](./docs/RELEASING_KREW.md): how to release new version of
  Krew.
- [Plugin Lifecycle](./docs/PLUGIN_LIFECYCLE.md): how Krew installs/upgrades
  plugins and itself. (Not necessarily up-to-date, but it can give a good idea
  about how Krew works under the covers.)
- [Krew Architecture](./docs/KREW_ARCHITECTURE.md): architectural decisions
  behind designing initial versions of Krew. (Not up-to-date.)
- [Krew Logo](./docs/KREW_LOGO.md): our logo and branding assets.

Visit [`./docs`](./docs) for all documentation.

## Roadmap

Please check out the [Issue
Tracker](https://github.com/kubernetes-sigs/krew/issues) to see the plan of
record for new features and changes.

## Community

### Bug reports

* If you have a problem with the Krew itself, please file an
  issue in this repository.
* If you're having a problem with a particular plugin's installation or
  upgrades, file an issue at [krew-index][index] repository.
* If you're having an issue with an installed plugin, file an issue for the
  repository the plugin's source code is hosted at.

### Communication channels

* Slack: [#krew](https://kubernetes.slack.com/messages/krew) or
  [#sig-cli](https://kubernetes.slack.com/messages/sig-cli)
* [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-sig-cli)
* [Kubernetes Community site](https://kubernetes.io/community/)

### Contributing

Interested in contributing to Krew? Please refer to our
[Contributing Guidelines](./docs/CONTRIBUTOR_GUIDE.md) for more details.

### Code of Conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code
of Conduct](https://github.com/kubernetes-sigs/krew/blob/master/code-of-conduct.md).

[index]:https://github.com/kubernetes-sigs/krew-index
[list]: https://krew.sigs.k8s.io/plugins/
