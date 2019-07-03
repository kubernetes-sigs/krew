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

[[ -n "${DEBUG:-}" ]] && set -x

export GO111MODULE=on
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


KREW_BINARY="${1:-$KREW_BINARY_DEFAULT}" # needed for `kubectl krew` in tests
if [[ ! -e "${KREW_BINARY}" ]]; then
  echo >&2 "Could not find $KREW_BINARY. You need to build krew for ${goos}/${goarch} before running the integration tests."
  exit 1
fi
krew_binary_realpath="$(readlink -f "${KREW_BINARY}")"
if [[ ! -x "${krew_binary_realpath}" ]]; then
  echo >&2 "krew binary at ${krew_binary_realpath} is not an executable"
  exit 1
fi
KREW_BINARY="${krew_binary_realpath}"
export KREW_BINARY

go test -test.v sigs.k8s.io/krew/integration_test
