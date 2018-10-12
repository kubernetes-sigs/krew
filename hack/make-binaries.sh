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
version_pkg="github.com/GoogleContainerTools/krew/pkg/version"

rm -rf out/

# Builds
echo >&2 "Building binaries for: ${OSARCH:-$DEFAULT_OSARCH}"
git_rev="${SHORT_SHA:-$(git rev-parse --short HEAD)}"
git_tag="${TAG_NAME:-$(git describe --tags --dirty --always)}"
echo >&2 "(Stamping with git tag=${git_tag} rev=${git_rev})"

env CGO_ENABLED=0 gox -osarch="${OSARCH:-$DEFAULT_OSARCH}" \
  -tags netgo \
  -ldflags="-w -X ${version_pkg}.gitCommit=${git_rev} \
    -X ${version_pkg}.gitTag=${git_tag}" \
  -output="out/bin/krew-{{.OS}}_{{.Arch}}" \
  ./cmd/krew/...
