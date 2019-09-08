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
        $krew install --manifest=out/krew.yaml --archive=out/krew.zip
    ```

### Release a new version

Krew follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).
Krew tags versions starting with `v`. Example: `v0.2.0-rc.1`.

1. **Decide on a version number:** set it to `$TAG` variable:

    ```sh
    TAG=v0.2.0-rc.1 # <- change this
    ```

1. **Update installation instructions:** Version number is hardcoded in
   `README.md`.

1. **Commit the changes back:**

       git commit -am "Release ${TAG:?TAG required}"

1. **Tag the release:**

    ```sh
    release_notes="$(TAG=$TAG hack/make-release-notes.sh)"
    git tag -a "${TAG:?TAG required}" -m "${release_notes}"
    ```

1. **Verify the release instructions:**

       git show "${TAG:?TAG required}"

1. **Push the tag:**

       git push --follow-tags

    Due to branch restrictions on GitHub preventing pushing to a branch
    directly, this command may require `-f`.

1. **Verify on Releases tab on GitHub**

1. **Make the new version available on krew index:** Copy the `krew.yaml` from
   the release artifacts and make a pull request to
   [krew-index](https://github.com/kubernetes-sigs/krew-index/) repository.
   This will make the plugin available to upgrade for users using older versions
   of krew.

## Release artifacts

When a tag is pushed to the repository, Travis CI will make a release on GitHub
and upload the release artifacts as files on the release.
