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

# Disallow usage of ioutil.TempDir in tests in favor of testutil.
out="$(grep --include '*_test.go' --exclude-dir 'vendor/' -EIrn 'ioutil.\TempDir' || true)"
if [[ -n "$out" ]]; then
  echo >&2 "You used ioutil.TempDir in tests, use 'testutil.NewTempDir()' instead:"
  echo >&2 "$out"
  exit 1
fi

out="$(grep --include '*.go' \
            --exclude "*_test.go" \
            --exclude 'constants.go' \
            --exclude-dir 'vendor/' \
            -EIrn '\.yaml"' || true)"
if [[ -n "$out" ]]; then
  echo >&2 'You used ".yaml" in production, use constants.ManifestExtension instead:'
  echo >&2 "$out"
  exit 1
fi
