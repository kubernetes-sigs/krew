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

# This script starts a development container by mounting the krew binary from
# the local filesystem.

set -euo pipefail
SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
log() { echo >&2 "$*"; }
log_ok() { log "$(tput setaf 2)$*$(tput sgr0)"; }
log_fail() { log "$(tput setaf 1)$*$(tput sgr0)"; }
image="krew:sandbox"

krew_bin="${SCRIPTDIR}/../out/bin/krew-linux_amd64"
if [[ ! -f "${krew_bin}" ]]; then
  log "Building the ${krew_bin}."
  env OSARCH="linux/amd64" "${SCRIPTDIR}/make-binaries.sh"
else
  log_ok "Using existing ${krew_bin}."
fi

docker build -f "${SCRIPTDIR}/sandboxed.Dockerfile" -q \
  --tag "${image}" "${SCRIPTDIR}/.."
log_ok "Sandbox image '${image}' built successfully."

kubeconfig="${KUBECONFIG:-$HOME/.kube/config}"
if [[ ! -f "${kubeconfig}" ]]; then
  log_fail "Warning: kubeconfig not found at ${kubeconfig}, using /dev/null"
  kubeconfig=/dev/null
fi

log_ok "Starting docker container with volume mounts:"
log "    kubeconfig=${kubeconfig}"
log "    kubectl-krew=${krew_bin}"
log_ok "You can rebuild with the following command without restarting the container:"
log "    env OSARCH=linux/amd64 hack/make-binaries.sh"
exec docker run --rm --tty --interactive \
  --volume "${krew_bin}:/usr/local/bin/kubectl-krew" \
  --volume "${kubeconfig}:/etc/kubeconfig" \
  --env KUBECONFIG=/etc/kubeconfig \
  --hostname krew \
  "${image}"
