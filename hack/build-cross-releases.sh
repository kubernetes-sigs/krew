#!/usr/bin/env bash

# Copyright 2018 Google LLC
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

set -e -o pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd "${SCRIPTDIR}/.."

# Builds
rm -rf out/
gox -os="darwin windows" -arch="amd64" \
  -ldflags="-X github.com/GoogleContainerTools/krew/pkg/version.gitCommit=$(git rev-parse --short HEAD) \
    -X github.com/GoogleContainerTools/krew/pkg/version.gitTag=$(git describe --tags --dirty --always)" \
  -output="out/build/krew-{{.OS}}_{{.Arch}}" \
  ./cmd/krew/...

gox -os="linux" -arch="arm amd64" \
  -ldflags="-X github.com/GoogleContainerTools/krew/pkg/version.gitCommit=$(git rev-parse --short HEAD) \
    -X github.com/GoogleContainerTools/krew/pkg/version.gitTag=$(git describe --tags --dirty --always)" \
  -output="out/build/krew-{{.OS}}_{{.Arch}}" \
  ./cmd/krew/...

go install github.com/GoogleContainerTools/krew/cmd/krew-manifest

(
  cd out/build/
  mkdir unix
  krew-manifest generate -o unix
  mkdir windows
  krew-manifest generate -o windows --windows=true
)

zip -X -q -r out/krew.zip out/build

KREW_HASH="$(shasum -a 256 out/krew.zip | awk '{print $1;}')"
echo "Computed Hash: ${KREW_HASH}"
echo "${KREW_HASH}" > out/krew-zip.sha256
