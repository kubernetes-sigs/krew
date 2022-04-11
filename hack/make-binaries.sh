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

# This script builds krew binaries for all supported platforms (or those os/arch
# combinations specified via OSARCH variable).

set -e -o pipefail
SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! command -v "gox" &>/dev/null; then
  echo >&2 "gox not installed in PATH, run hack/install-gox.sh."
  exit 1
fi

supported_platforms="darwin/amd64 darwin/arm64 windows/amd64\
 linux/amd64 linux/arm linux/arm64 linux/s390x"
version_pkg="sigs.k8s.io/krew/internal/version"

cd "${SCRIPTDIR}/.."
rm -rf -- "out/"

# Builds
echo >&2 "Building binaries for: ${OSARCH:-$supported_platforms}"
git_rev="${SHORT_SHA:-$(git rev-parse --short HEAD)}"
git_tag="${TAG_NAME:-$(git describe --tags --dirty --always)}"
echo >&2 "(Stamping with git tag=${git_tag} rev=${git_rev})"

env CGO_ENABLED=0 gox -osarch="${OSARCH:-$supported_platforms}" \
  -tags netgo \
  -mod readonly \
  -ldflags="-w -X ${version_pkg}.gitCommit=${git_rev} \
    -X ${version_pkg}.gitTag=${git_tag}" \
  -output="out/bin/krew-{{.OS}}_{{.Arch}}" \
  ./cmd/krew/...
