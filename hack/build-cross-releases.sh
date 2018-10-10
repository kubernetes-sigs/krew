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

DEFAULT_OSARCH="darwin/amd64 windows/amd64 linux/amd64 linux/arm"
KREW_ARCHIVE="krew.zip"

# Builds
rm -rf out/
echo "Building releases: ${OSARCH:-$DEFAULT_OSARCH}"
gox -osarch="${OSARCH:-$DEFAULT_OSARCH}" \
  -ldflags="-X github.com/GoogleContainerTools/krew/pkg/version.gitCommit=$(git rev-parse --short HEAD) \
    -X github.com/GoogleContainerTools/krew/pkg/version.gitTag=$(git describe --tags --dirty --always)" \
  -output="out/bin/krew-{{.OS}}_{{.Arch}}" \
  ./cmd/krew/...

(
  set -x;
  zip -X -q -r --verbose out/krew.zip out/bin
)

zip_checksum="$(shasum -a 256 "out/${KREW_ARCHIVE}" | awk '{print $1;}')"
echo "${KREW_ARCHIVE} checksum: ${zip_checksum}"
echo "${zip_checksum}" > "out/${KREW_ARCHIVE}".sha256

# Copy and process krew manifest
cp ./hack/krew.yaml ./out/krew.yaml
tag="$(git describe --tags HEAD)"
sed -i "s/KREW_ZIP_CHECKSUM/${zip_checksum}/g" ./out/krew.yaml
sed -i "s/KREW_TAG/${tag}/g" ./out/krew.yaml
echo "Manifest processed."
