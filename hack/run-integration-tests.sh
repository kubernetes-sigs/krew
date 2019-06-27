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
goos="$(go env GOOS)"
goarch="$(go env GOARCH)"
KREW_BINARY_DEFAULT="${BINDIR}/krew-${goos}_${goarch}"

if [[ "$#" -gt 0 && ( "$1" = '-h' || "$1" = '--help' ) ]]; then
  cat <<EOF
Usage:
  $0 krew  # uses the given krew binary for running integration tests
  $0       # assumes a pre-built krew binary at $KREW_BINARY_DEFAULT
EOF
  exit 0
fi

KREW_BINARY=$(readlink -f "${1:-$KREW_BINARY_DEFAULT}")  # needed for `kubectl krew` in tests
if [[ ! -x "${KREW_BINARY}" ]]; then
  echo "Did not find $KREW_BINARY. You need to build krew FOR ${goos}/${goarch} before running the integration tests."
  exit 1
fi
export KREW_BINARY

"${SCRIPTDIR}/ensure-kubectl-installed.sh"
go test -test.v sigs.k8s.io/krew/test