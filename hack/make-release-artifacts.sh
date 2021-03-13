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

set -e -o pipefail

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPTDIR}/.."

bin_dir="out/bin"
if [[ ! -d "${bin_dir}" ]]; then
  echo >&2 "Binaries are not built (${bin_dir}), run hack/make-binaries.sh"
  exit 1
fi

checksum_cmd="shasum -a 256"
if hash sha256sum 2>/dev/null; then
  checksum_cmd="sha256sum"
fi
checksum_sed=""

while IFS= read -r -d $'\0' f; do
  archive_dir="$(mktemp -d)"
  cp "$f" "${archive_dir}"
  cp -- "${SCRIPTDIR}/../LICENSE" "${archive_dir}"
  name="$(basename "$f" .exe)"
  archive="${name}.tar.gz"
  echo >&2 "Creating ${archive} archive."
  (
    cd "${archive_dir}"
    # consistent timestamps for files in archive dir to ensure consistent checksums
    TZ=UTC touch -t "0001010000" ./*
    tar --use-compress-program "gzip --no-name" -cvf "${SCRIPTDIR}/../out/${archive}" ./*
  )

  # create sumfile
  sumfile="out/${archive}.sha256"
  checksum="$(eval "${checksum_cmd[@]}" "out/${archive}" | awk '{print $1;}')"
  echo >&2 "${archive} checksum: ${checksum}"
  echo "${checksum}" >"${sumfile}"
  echo >&2 "Written ${sumfile}."

  # prepare krew manifest sed
  checksum_sed="${checksum_sed};s/$(tr "[[:lower:]-]" "[[:upper:]_]" <<<${name})_CHECKSUM/${checksum}/"

done < <(find "${bin_dir}" -type f -print0)

# create a out/krew.exe convenience copy
if [[ -x "./${bin_dir}/krew-windows_amd64.exe" ]]; then
  krew_exe="krew.exe"
  cp -- "./${bin_dir}/krew-windows_amd64.exe" "./out/${krew_exe}"
  exe_sumfile="out/krew.exe.sha256"
  exe_checksum="$(eval "${checksum_cmd[@]}" "out/${krew_exe}" | awk '{print $1;}')"
  echo >&2 "${krew_exe} checksum: ${exe_checksum}"
  echo "${exe_checksum}" >"${exe_sumfile}"
  echo >&2 "Written ${exe_sumfile}."
fi

# Copy and process krew manifest
git_describe="$(git describe --tags --dirty --always)"
if [[ ! "${git_describe}" =~ v.* ]]; then
  # if tag cannot be inferred (e.g. CI/CD), still provide a valid
  # version field for krew.yaml
  git_describe="v0.0.0-detached+${git_describe}"
fi
krew_version="${TAG_NAME:-$git_describe}"
sed "${checksum_sed};s/KREW_TAG/${krew_version}/g" ./hack/krew.yaml >./out/krew.yaml
echo >&2 "Written out/krew.yaml."
