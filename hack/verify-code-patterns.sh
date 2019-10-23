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

# Appropriately set grep
grep="grep"
if [[ $(uname) == "Darwin" ]]; then
  out="$(command -v ggrep)"
  if [[ -n $out ]]; then
    grep="ggrep"
  else
    echo >&2 "GNU grep is required for MacOS users to successfully run local tests; it can be found on homebrew as 'grep'"
  fi
fi

# Disallow usage of ioutil.TempDir in tests in favor of testutil.
out="$(${grep} --include '*_test.go' --exclude-dir 'vendor/' -EIrn 'ioutil.\TempDir' || true)"
if [[ -n "$out" ]]; then
  echo >&2 "You used ioutil.TempDir in tests, use 'testutil.NewTempDir()' instead:"
  echo >&2 "$out"
  exit 1
fi

# use code constant for ".yaml"
out="$(${grep} --include '*.go' \
  --exclude "*_test.go" \
  --exclude 'constants.go' \
  --exclude-dir 'vendor/' \
  -EIrn '\.yaml"' || true)"
if [[ -n "$out" ]]; then
  echo >&2 'You used ".yaml" in production, use constants.ManifestExtension instead:'
  echo >&2 "$out"
  exit 1
fi

# Do not use glog in test code
out="$(${grep} --include '*_test.go' --exclude-dir 'vendor/' -EIrn '[kg]log\.' || true)"
if [[ -n "$out" ]]; then
  echo >&2 "You used glog in tests, use 't.Logf' instead:"
  echo >&2 "$out"
  exit 1
fi

# Do not initialize index.{Plugin,Platform} structs in test code.
out="$(${grep} --include '*_test.go' --exclude-dir 'vendor/' -EIrn '[^]](index\.)(Plugin|Platform){' || true)"
if [[ -n "$out" ]]; then
  echo >&2 "Do not use index.Platform or index.Plugin structs directly in tests,"
  echo >&2 "use testutil.NewPlugin() or testutil.NewPlatform() instead:"
  echo >&2 "-----"
  echo >&2 "$out"
  exit 1
fi
