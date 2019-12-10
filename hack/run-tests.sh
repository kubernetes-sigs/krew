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

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

color_red="$(tput setaf 1)"
color_green="$(tput setaf 2)"
color_blue="$(tput setaf 4)"
color_reset="$(tput sgr0)"

print_with_color() {
  echo "${1}${*:2}${color_reset}"
}

print_status() {
  local result=$? # <- this must be the first action
  if [[ $result == 0 ]]; then
    print_with_color "$color_green" 'SUCCESS'
  else
    print_with_color "$color_red" 'FAILURE'
  fi
}
trap print_status EXIT

print_with_color "$color_blue" 'Checking boilerplate'
"$SCRIPTDIR"/verify-boilerplate.sh

print_with_color "$color_blue" 'Running tests'
go test -short -race sigs.k8s.io/krew/...

print_with_color "$color_blue" 'Running linter'
"$SCRIPTDIR"/run-lint.sh

print_with_color "$color_blue" 'Check code patterns'
"$SCRIPTDIR"/verify-code-patterns.sh
