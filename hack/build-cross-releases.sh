#!/bin/bash

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
gox -os="linux darwin windows" -arch="amd64" -output="out/build/krew-{{.OS}}" ./cmd/krew/...
go install github.com/GoogleContainerTools/krew/cmd/krew-manifest

cd out/build/
mkdir unix
krew-manifest generate -o unix
mkdir windows
krew-manifest generate -o windows --windows=true
cd ..

# reproducible
rm -f krew.zip
zip -X -q -r krew.zip build

KREW_HASH="$(shasum -a 256 ./krew.zip|awk '{print $1;}')"
echo "${KREW_HASH}"
echo "${KREW_HASH}" > krew-zip.sha256
