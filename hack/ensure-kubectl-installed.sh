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

install_kubectl_if_needed() {
  if hash kubectl 2>/dev/null; then
    echo >&2 "using kubectl from the host system and not reinstalling"
  else
    local bin_dir
    bin_dir="$(go env GOPATH)/bin"
    local -r kubectl_version='v1.14.2'
    local -r kubectl_path="${bin_dir}/kubectl"
    local goos goarch kubectl_url
    goos="$(go env GOOS)"
    goarch="$(go env GOARCH)"
    kubectl_url="https://dl.k8s.io/release/${kubectl_version}/bin/${goos}/${goarch}/kubectl"

    echo >&2 "kubectl not detected in environment, downloading ${kubectl_url}"
    mkdir -p "${bin_dir}"
    curl --fail --show-error --silent --location --output "$kubectl_path" "${kubectl_url}"
    chmod +x "$kubectl_path"
    echo >&2 "installed kubectl to ${kubectl_path}"
  fi
}

install_kubectl_if_needed
