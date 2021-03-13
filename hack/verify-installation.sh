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

# This script verifies that a krew build can be installed to a system using
# itself as the documented installation method.

set -euo pipefail

[[ -n "${DEBUG:-}" ]] && set -x

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
build_dir="${SCRIPTDIR}/../out"
goos="$(go env GOOS)"
goarch="$(go env GOARCH)"

krew_manifest="${build_dir}/krew.yaml"
if [[ ! -f "${krew_manifest}" ]]; then
  echo >&2 "Could not find manifest ${krew_manifest}."
  echo >&2 "Did you run hack/make-all.sh?"
  exit 1
fi

krew_archive="${build_dir}/krew-${goos}_${goarch}.tar.gz"
if [[ ! -f "${krew_archive}" ]]; then
  echo >&2 "Could not find archive ${krew_archive}."
  echo >&2 "Did you run hack/make-all.sh?"
  exit 1
fi

temp_dir="$(mktemp -d)"
trap 'rm -rf -- "${temp_dir}"' EXIT
echo >&2 "Extracting krew from tarball."
tar zxf "${krew_archive}" -C "${temp_dir}"
krew_binary="${temp_dir}/krew-${goos}_${goarch}"

krew_root="$(mktemp -d)"
trap 'rm -rf -- "${krew_root}"' EXIT
system_path="/usr/local/bin:/usr/bin:/bin:/usr/local/sbin:/usr/sbin:/sbin"

echo >&2 "Installing the krew build to a temporary directory."
env -i KREW_ROOT="${krew_root}" \
  "${krew_binary}" install \
  --manifest="${krew_manifest}" \
  --archive "${krew_archive}"

echo >&2 "Verifying krew installation (symlink)."
env -i PATH="${krew_root}/bin:${system_path}" /bin/bash -c \
  "which kubectl-krew 1>/dev/null"
