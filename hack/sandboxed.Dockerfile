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

FROM golang:alpine as builder

WORKDIR /go/src/sigs.k8s.io/krew

ENV KUBECTL_VERSION v1.14.2
RUN apk add --no-cache curl && \
    curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl  && \
  chmod +x /usr/bin/kubectl

# build binary
COPY . .
RUN go build -tags netgo -ldflags "-s -w" ./cmd/krew

# production image
FROM alpine
RUN apk --no-cache add git && \
    ln -s /usr/bin/krew /usr/bin/kubectl-krew

# initialize index
RUN mkdir -p /root/.krew/index && \
    git clone https://github.com/kubernetes-sigs/krew-index /root/.krew/index

COPY --from=builder /go/src/sigs.k8s.io/krew/krew /usr/bin/kubectl /usr/bin/
