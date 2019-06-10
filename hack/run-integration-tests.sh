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

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BINDIR="${SCRIPTDIR}/../out/bin"
GOOS="$(go env GOOS)"
GOARCH="$(go env GOARCH)"
KREW_BINARY_DEFAULT="$BINDIR/krew-${GOOS}_${GOARCH}"

if [[ "$#" -gt 0 && ( "$1" = '-h' || "$1" = '--help' ) ]]; then
  cat <<EOF
Usage:
  $0 krew  # uses the given krew binary for running integration tests
  $0       # assumes a pre-built krew binary at $KREW_BINARY_DEFAULT
EOF
  exit 0
fi

install_kubectl_if_needed() {
  if hash kubectl 2>/dev/null; then
    echo 'using kubectl from the host system'
  else
    # install kubectl
    local -r KUBECTL_VERSION='v1.14.2'
    local -r KUBECTL_BINARY="$BINDIR/kubectl"
    curl -fSsLo "$KUBECTL_BINARY" https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl
    chmod +x "$KUBECTL_BINARY"
    export PATH="$BINDIR:$PATH"
  fi
}

install_kubectl_if_needed
KREW_BINARY=$(readlink -f "${1:-$KREW_BINARY_DEFAULT}")  # needed for `kubectl krew` in tests
export KREW_BINARY

go test -v ./...
