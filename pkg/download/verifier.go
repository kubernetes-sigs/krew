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

package download

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// Verifier can check a reader against it's correctness.
type Verifier interface {
	io.Writer
	Verify() error
}

var _ Verifier = sha256Verifier{}

type sha256Verifier struct {
	hash.Hash
	wantedHash []byte
}

// NewSha256Verifier creates a Verifier that tests against the given hash.
func NewSha256Verifier(hash string) Verifier {
	raw, _ := hex.DecodeString(hash)
	return sha256Verifier{
		Hash:       sha256.New(),
		wantedHash: raw,
	}
}

func (v sha256Verifier) Verify() error {
	glog.V(1).Infof("Compare sha256 (%s) signed version", hex.EncodeToString(v.wantedHash))
	if bytes.Equal(v.wantedHash, v.Sum(nil)) {
		return nil
	}
	return errors.Errorf("checksum does not match, want: %x, got %x", v.wantedHash, v.Sum(nil))
}

var _ Verifier = trueVerifier{}

type trueVerifier struct{ io.Writer }

// NewInsecureVerifier returns a Verifier that always verifies to true.
func NewInsecureVerifier() Verifier { return trueVerifier{ioutil.Discard} }
func (trueVerifier) Verify() error  { return nil }
