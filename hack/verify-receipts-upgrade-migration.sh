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

# This script tests "krew system receipts-migration" which was introduced for
# migrating krew 0.2.x to krew 0.3.x.
#
# TODO(ahmetb,corneliusweig) remove at/after krew 0.4.x when receipts-migration
# is no longer supported.

set -euo pipefail

[[ -n "${DEBUG:-}" ]] && set -x

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINDIR="${SCRIPTDIR}/../out/bin"
goos="$(go env GOOS)"
goarch="$(go env GOARCH)"

install_krew_0_2_1() {
  # from https://github.com/kubernetes-sigs/krew/blob/v0.2.1/README.md
  krew_root="${1}"
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "${temp_dir}"' RETURN
  (
    cd "${temp_dir}"
    curl -fsSLO "https://storage.googleapis.com/krew/v0.2.1/krew.{tar.gz,yaml}"
    tar zxf krew.tar.gz
    env KREW_ROOT="${krew_root}" ./krew-"${goos}_amd64" install \
      --manifest="./krew.yaml" --archive="./krew.tar.gz"
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

verify_plugin_installed() {
  krew_root="${1}"
  plugin="${2}"
  local plugin_bin_name
  plugin_bin_name="kubectl-${plugin//-/_}"

  [[ -x "${krew_root}/bin/${plugin_bin_name}" ]]
}

verify_plugin_receipt() {
  krew_root="${1}"
  plugin="${2}"
  [[ -f "${krew_root}/receipts/${plugin}.yaml" ]]
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
  install_krew_0_2_1 "${krew_root}"
  install_plugin "${krew_root}" "get-all"
  echo >&2 "Swapping krew binary."
  patch_krew_bin "${krew_root}" "${new_krew}"
  run_krew "${krew_root}" version && (
    echo >&2 "[FAIL] expected 'krew version' to fail with upgrade message"
    exit 1
  )
  echo >&2 "Performing migration."
  run_krew "${krew_root}" system receipts-upgrade
  verify_plugin_installed "${krew_root}" get-all || (
    echo >&2 "get-all plugin is not linked"
    exit 1
  )
  verify_plugin_receipt "${krew_root}" get-all || (
    echo >&2 "get-all plugin receipt missing"
    exit 1
  )
  run_krew "${krew_root}" list || (
    echo >&2 "krew list is failing"
    exit 1
  )
  run_krew "${krew_root}" search || (
    echo >&2 "krew search is failing"
    exit 1
  )
  run_krew "${krew_root}" uninstall get-all || (
    echo >&2 "get-all cannot be uninstalled"
  )
  verify_plugin_installed "${krew_root}" krew || (
    echo >&2 "krew binary is missing"
    exit 1
  )
  verify_plugin_receipt "${krew_root}" krew || (
    echo >&2 "krew plugin receipt is not copied over"
    exit 1
  )
  install_plugin "${krew_root}" "who-can" # install still works after migration
  verify_plugin_installed "${krew_root}" who-can || (
    echo >&2 "who-can plugin is not linked"
    exit 1
  )
  verify_plugin_receipt "${krew_root}" who-can || (
    echo >&2 "who-can plugin receipt missing"
    exit 1
  )
}

main
