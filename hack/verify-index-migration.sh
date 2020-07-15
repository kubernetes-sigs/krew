#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
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

# This script tests the automatic index migration which was added for
# migrating krew 0.3.x to krew 0.4.x.
#
# TODO(ahmetb,corneliusweig,chriskim06) remove at/after krew 0.5.x when
# index-migration is no longer supported.

set -euo pipefail

[[ -n "${DEBUG:-}" ]] && set -x

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINDIR="${SCRIPTDIR}/../out/bin"
goos="$(go env GOOS)"
goarch="$(go env GOARCH)"

install_krew_0_3_4() {
  krew_root="${1}"
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "${temp_dir}"' RETURN
  (
  set -x; cd "$(mktemp -d)" &&
    curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/krew.{tar.gz,yaml}" &&
    tar zxvf krew.tar.gz &&
    KREW=./krew-"$(uname | tr '[:upper:]' '[:lower:]')_amd64" &&
    "$KREW" install --manifest=krew.yaml --archive=krew.tar.gz &&
    "$KREW" update
  )
}

install_plugin() {
  krew_root="${1}"
  plugin="${2}"

  run_krew "${1}" install "${plugin}" 1>/dev/null
}

# patch_krew_bin replaces the installed krew 0.2.x binary with the specified
patch_krew_bin() {
  krew_root="${1}"
  new_binary="${2}"

  local old_binary
  old_binary="$(readlink -f "${krew_root}/bin/kubectl-krew")"
  cp -f "${new_binary}" "${old_binary}"
}

# run_krew runs 'krew' with the specified KREW_ROOT and arguments.
run_krew() {
  krew_root="${1}"
  shift

  env KREW_ROOT="${krew_root}" \
    PATH="${krew_root}/bin:$PATH" \
    kubectl krew "$@"
}

# run_krew runs 'krew' with the specified KREW_ROOT and arguments.
run_krew_with_multi_index_flag() {
  krew_root="${1}"
  shift

  env KREW_ROOT="${krew_root}" \
    PATH="${krew_root}/bin:$PATH" \
    X_KREW_ENABLE_MULTI_INDEX=1 \
    kubectl krew "$@"
}

verify_index_migrated() {
  krew_root="${1}"
  [[ -d "${krew_root}/index/default" ]]
}

main() {
  new_krew="${BINDIR}/krew-${goos}_${goarch}"
  if [[ ! -e "${new_krew}" ]]; then
    echo >&2 "Could not find ${new_krew}."
    exit 1
  fi

  krew_root="$(mktemp -d)"
  trap 'rm -rf "${krew_root}"' RETURN

  echo >&2 "Test directory: ${krew_root}"
  install_krew_0_3_4 "${krew_root}"
  install_plugin "${krew_root}" "get-all"
  echo >&2 "Swapping krew binary"
  patch_krew_bin "${krew_root}" "${new_krew}"
  run_krew_with_multi_index_flag "${krew_root}" list || (
    echo >&2 "krew list is failing"
    exit 1
  )
  verify_index_migrated "${krew_root}" || (
    echo >&2 "index was not migrated"
    ls -la "${krew_root}/index"
    exit 1
  )
}

main
