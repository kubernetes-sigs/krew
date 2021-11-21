// Copyright 2019 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"github.com/fatih/color"
	"sigs.k8s.io/krew/pkg/constants"
)

const securityNoticeFmt = `You installed plugin %q from the krew-index plugin repository.
   These plugins are not audited for security by the Krew maintainers.
   Run them at your own risk.`

var stderr = color.Error

func PrintSecurityNotice(plugin string) {
	if plugin == constants.KrewPluginName {
		return // do not warn for krew itself
	}
	PrintWarning(stderr, securityNoticeFmt+"\n", plugin)
}
