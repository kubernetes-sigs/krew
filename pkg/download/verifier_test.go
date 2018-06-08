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

package download

import (
	"bytes"
	"io"
	"testing"
)

func TestSha256Verifier(t *testing.T) {
	type args struct {
		hash string
	}
	tests := []struct {
		name      string
		args      args
		write     []byte
		wantError bool
	}{
		{
			name: "test okay hash",
			args: args{
				hash: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			},
			write:     []byte("hello world"),
			wantError: false,
		},
		{
			name: "test wrong hash",
			args: args{
				hash: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			},
			write:     []byte("HELLO WORLD"),
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSha256Verifier(tt.args.hash)
			io.Copy(v, bytes.NewReader(tt.write))
			if err := v.Verify(); (err != nil) != tt.wantError {
				t.Errorf("NewSha256Verifier().Write(%x).Verify() = %v, want %v", tt.write, err, tt.wantError)
				return
			}
		})
	}
}

func TestTrueVerifier(t *testing.T) {
	tests := []struct {
		name      string
		write     []byte
		wantError bool
	}{
		{
			name:      "test okay hash",
			write:     []byte("hello world"),
			wantError: false,
		},
		{
			name:      "test wrong hash",
			write:     []byte("HELLO WORLD"),
			wantError: false,
		},
		{
			name:      "test empty hash",
			write:     []byte{},
			wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewTrueVerifier()
			io.Copy(v, bytes.NewReader(tt.write))
			if err := v.Verify(); (err != nil) != tt.wantError {
				t.Errorf("NewTrueVerifier().Write(%x).Verify() = %v, want %v", tt.write, err, tt.wantError)
				return
			}
		})
	}
}
