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

gopath="$(go env GOPATH)"

if ! [[ -x "$gopath/bin/golangci-lint" ]]; then
  echo >&2 'Installing golangci-lint'
  curl --silent --fail --location \
    https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$gopath/bin" v1.57.2
fi

# configured by .golangci.yml
"$gopath/bin/golangci-lint" run

# install shfmt that ensures consistent format in shell scripts
if ! [[ -x "${gopath}/bin/shfmt" ]]; then
  echo >&2 'Installing shfmt'
  go install mvdan.cc/sh/v3/cmd/shfmt@v3.0.0
fi

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
shfmt_out="$($gopath/bin/shfmt -l -i=2 ${SCRIPTDIR})"
if [[ -n "${shfmt_out}" ]]; then
  echo >&2 "The following shell scripts need to be formatted, run: 'shfmt -w -i=2 ${SCRIPTDIR}'"
  echo >&2 "${shfmt_out}"
  exit 1
fi
