// Copyright © 2018 Google Inc.
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

package pathutil

import (
	"path/filepath"
	"strings"
)

// IsSubPath checks if the extending path is an extension of the basePath, it will return the extending path
// elements. Both paths have to be absolute or have the same root directory. The remaining path elements
func IsSubPath(subPath, path string) ([]string, bool) {
	basePieces := strings.Split(filepath.Clean(subPath), string(filepath.Separator))
	extendingPieces := strings.Split(filepath.Clean(path), string(filepath.Separator))

	// the binary has to be in the install path.
	if len(basePieces) > len(extendingPieces) {
		return nil, false
	}

	// Compare path pieces.
	for i, p := range basePieces {
		if extendingPieces[i] != p {
			return nil, false
		}
	}

	return extendingPieces[len(basePieces):], true
}
