# Releasing Krew

(This document is intended for maintainers of Krew only.)

### Build/Test the release locally

1. Build krew reelase assets locally:

       hack/make-all.sh

2. Try krew installation on each platform:

    ```sh
    krew=out/bin/krew-darwin_amd64 # assuming macOS amd64

    for osarch in darwin_amd64 darwin_arm64 linux_amd64 linux_arm linux_arm64 windows_amd64; do
      KREW_ROOT="$(mktemp -d --tmpdir krew-XXXXXXXXXX)" KREW_OS="${osarch%_*}" KREW_ARCH="${osarch#*_}" \
          $krew install --manifest=out/krew.yaml --archive="out/krew-${osarch}.tar.gz"
    done
    ```

### Release a new version

Krew follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).
Krew tags versions starting with `v`. Example: `v0.2.0-rc.1`.

1. **Decide on a version number:** set it to `$TAG` variable:

    ```sh
    TAG=v0.3.2-rc.1 # <- change this
    ```

1. **Create a release commit:**

       git commit -am "Release ${TAG:?TAG required}" --allow-empty

1. **Push PR and merge changes**: The repository hooks forbid direct pushes to
   master, so the changes from the previous step need to be pushed and merged
   as a regular PR.

       git push origin master

   (Only repository administrators can directly push to master branch.)

1. **Wait until the build succeeds:** Wait for CI to show green for the
   build of the commit you just pushed to master branch.

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

       git push origin "${TAG:?TAG required}"

1. **Verify on Releases tab on GitHub:** Make sure `krew.yaml`, `krew.tar.gz`
   and other release assets show up on "Releases" tab.

1. **Make the new version available on krew index:** Get the latest `krew.yaml` from

       curl -LO https://github.com/kubernetes-sigs/krew/releases/download/"${TAG:?TAG required}"/krew.yaml

   and make a pull request to
   [krew-index](https://github.com/kubernetes-sigs/krew-index/) repository.
   This will make the plugin available to upgrade for users using older versions
   of krew.

1. **Update krew-index CI**: The CI tests for `krew-index` repository relies on
   tools from main `krew` repository, and they should use the latest version.
   When there's a new version, update the `.travis.yml` in `krew-index` repo.

## Release artifacts

When a tag is pushed to the repository, GitHub workflow will make a release
on GitHub, and upload the release artifacts as files on the release.
