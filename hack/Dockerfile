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

# Used by cloudbuild*.yaml to build in Google Cloud Build.
FROM golang:1-alpine
RUN apk add -q --no-cache bash git zip
RUN go get github.com/mitchellh/gox

ENV PROJECT_ROOT /go/src/sigs.k8s.io/krew
RUN mkdir -p $(dirname $PROJECT_ROOT) && \
        ln -s /workspace $PROJECT_ROOT
WORKDIR $PROJECT_ROOT
ENTRYPOINT $PROJECT_ROOT/hack/make-all.sh
