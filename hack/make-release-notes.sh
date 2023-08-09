#!/usr/bin/env bash

# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script should be executed while tagging a commit.
# You can run this script while tagging the release as:
#     git tag -a v0.1 -m "$(TAG=v0.1 hack/make-release-notes.sh)"

set -euo pipefail

gopath="$(go env GOPATH)"

TAG="${TAG:?TAG environment variable must be set for this script}"
if ! [[ "$TAG" =~ v.* ]]; then
  echo >&2 "TAG must be in format v.*"
  exit 1
fi

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPTDIR}/.."

archive_dir="out"
if [[ ! -d "${archive_dir}" ]]; then
  echo >&2 "Archive dir is not created (${archive_dir}), run hack/make-all.sh"
  exit 1
fi

cd "${archive_dir}"

download_assets=()
for entry in *; do
  if [[ -f "${entry}" ]]; then
    download_assets[${#download_assets[@]}]="${entry}"
  fi
done
if [[ ${#download_assets[@]} == 0 ]]; then
  echo >&2 "Archives are not created, run hack/make-release-artifacts.sh"
  exit 1
fi

readme="https://github.com/kubernetes-sigs/krew/blob/${TAG}/README.md"
download_base="https://github.com/kubernetes-sigs/krew/releases/download"

# install release-notes tool if not present
if [[ ! -f "${gopath}/bin/release-notes" ]]; then
  echo >&2 'Installing release-notes tool...'
  go install github.com/corneliusweig/release-notes@v0.1.0
fi

echo "Installation"
echo "------------"
echo "To install this release, refer to the instructions at ${readme}."
echo
echo "Release Assets"
echo "--------------"
echo "Artifacts for this release can be downloaded from the following links."
echo "It is recommended to follow [installation instructions](${readme})"
echo "and not using these artifacts directly."
echo
for f in "${download_assets[@]}"; do
  echo "- $download_base/${TAG}/${f}"
done
echo
echo "Thanks to our contributors for helping out with ${TAG}:"
previous_version="$(git describe --tags --match 'v*' --abbrev=0 "${TAG}^")"
git log "${previous_version}..${TAG}" --format=%an |
  sort | uniq -c | sort -rn |
  sed -E 's,^(\s+[0-9]+\s),- ,g'
echo
echo "(krew ${TAG} was tagged on $(date -u).)"
echo
echo '<details>'
echo '<summary>Merged pull requests</summary>'
echo # this empty line is important for correct markdown rendering
# you can pass your github token with --token here if you run out of requests

"${gopath}/bin/release-notes" kubernetes-sigs krew --since "${previous_version}"
echo '</details>'
echo
