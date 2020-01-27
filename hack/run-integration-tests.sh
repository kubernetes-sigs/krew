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

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINDIR="${SCRIPTDIR}/../out/bin"
goos="$(go env GOOS)"
goarch="$(go env GOARCH)"
krew_binary_default="${BINDIR}/krew-${goos}_${goarch}"

if [[ "$#" -gt 0 && ("$1" == '-h' || "$1" == '--help') ]]; then
  cat <<EOF
Runs the integration tests against built krew binary.
Set KREW_BINARY to use a krew binary at a different location.
Positional arguments are passed to the underlying 'go test'.

Usage:
  $0
  $0 -test.v -test.run TestFoo
  env KREW_BINARY=[FILE] $0
EOF
  exit 0
fi

KREW_BINARY="${KREW_BINARY:-$krew_binary_default}" # needed for `kubectl krew` in tests

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

go test sigs.k8s.io/krew/integration_test "$@"
