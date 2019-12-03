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

krew_tar_archive="krew.tar.gz"
krew_exe="krew.exe"

# copy license
cp -- "${SCRIPTDIR}/../LICENSE" "./${bin_dir}"

# create a out/krew.exe convenience copy
if [[ -x "./${bin_dir}/krew-windows_amd64.exe" ]]; then
  cp -- "./${bin_dir}/krew-windows_amd64.exe" "./out/krew.exe"
fi

# consistent timestamps for files in bindir to ensure consistent checksums
while IFS= read -r -d $'\0' f; do
  echo "modifying atime/mtime for $f"
  TZ=UTC touch -at "0001010000" "$f"
  TZ=UTC touch -mt "0001010000" "$f"
done < <(find "${bin_dir}" -print0)

echo >&2 "Creating ${krew_tar_archive} archive."
(
  cd "${bin_dir}"
  tar --use-compress-program "gzip --no-name" -cvf "${SCRIPTDIR}/../out/${krew_tar_archive}" ./*
)

checksum_cmd="shasum -a 256"
if hash sha256sum 2>/dev/null; then
  checksum_cmd="sha256sum"
fi

tar_sumfile="out/${krew_tar_archive}.sha256"
tar_checksum="$(eval "${checksum_cmd[@]}" "out/${krew_tar_archive}" | awk '{print $1;}')"
echo >&2 "${krew_tar_archive} checksum: ${tar_checksum}"
echo "${tar_checksum}" >"${tar_sumfile}"
echo >&2 "Written ${tar_sumfile}."

exe_sumfile="out/krew.exe.sha256"
exe_checksum="$(eval "${checksum_cmd[@]}" "out/${krew_exe}" | awk '{print $1;}')"
echo >&2 "${krew_exe} checksum: ${exe_checksum}"
echo "${exe_checksum}" >"${exe_sumfile}"
echo >&2 "Written ${exe_sumfile}."

# Copy and process krew manifest
git_describe="$(git describe --tags --dirty --always)"
if [[ ! "${git_describe}" =~ v.* ]]; then
  # if tag cannot be inferred (e.g. CI/CD), still provide a valid
  # version field for krew.yaml
  git_describe="v0.0.0-detached.${git_describe}"
fi
krew_version="${TAG_NAME:-$git_describe}"
cp ./hack/krew.yaml ./out/krew.yaml
sed -i "s/KREW_TAR_CHECKSUM/${tar_checksum}/g" ./out/krew.yaml
sed -i "s/KREW_TAG/${krew_version}/g" ./out/krew.yaml
echo >&2 "Written out/krew.yaml."
