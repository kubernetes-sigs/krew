# Releasing Krew

(This document is intended for maintainers of Krew only.)

### Build/Test the release locally

1. Build krew reelase assets locally:

       hack/make-all.sh

2. Try krew installation on each platform:

    ```sh
    krew=out/bin/krew-darwin_amd64 # assuming macOS

    KREW_ROOT="$(mktemp -d)" KREW_OS=darwin \
        $krew install --manifest=out/krew.yaml --archive=out/krew.tar.gz && \
    KREW_ROOT="$(mktemp -d)" KREW_OS=linux \
        $krew install --manifest=out/krew.yaml --archive=out/krew.tar.gz && \
    KREW_ROOT="$(mktemp -d)" KREW_OS=windows \
        $krew install --manifest=out/krew.yaml --archive=out/krew.tar.gz
    ```

### Release a new version

Krew follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).
Krew tags versions starting with `v`. Example: `v0.2.0-rc.1`.

1. **Decide on a version number:** set it to `$TAG` variable:

    ```sh
    TAG=v0.3.2-rc.1 # <- change this
    ```

1. **Update installation instructions:** Version number is hardcoded in
   `README.md`.

1. **Commit the changes back:**

       git commit -am "Release ${TAG:?TAG required}"

1. **Push PR and merge changes**: The repository hooks forbid direct pushes to
   master, so the changes from the previous step need to be pushed and merged
   as a regular PR.

1. **Tag the release:**

    ```sh
    git fetch origin
    git reset --hard origin/master    # when the previous merge is done
    release_notes="$(TAG=$TAG hack/make-release-notes.sh)"
    git tag -a "${TAG:?TAG required}" -m "${release_notes}"
    ```

1. **Verify the release instructions:**

       git show "${TAG:?TAG required}"

1. **Push the tag:**

       git push --tags

1. **Verify on Releases tab on GitHub**

1. **Make the new version available on krew index:** Get the latest `krew.yaml` from

       curl -LO https://github.com/kubernetes-sigs/krew/releases/download/v0.3.2-rc.1/krew.yaml

   and make a pull request to
   [krew-index](https://github.com/kubernetes-sigs/krew-index/) repository.
   This will make the plugin available to upgrade for users using older versions
   of krew.

## Release artifacts

When a tag is pushed to the repository, Travis CI will make a release on GitHub
and upload the release artifacts as files on the release.
