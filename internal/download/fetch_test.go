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
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestHTTPFetcher_Get(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name    string
		h       HTTPFetcher
		args    args
		want    io.ReadCloser
		wantErr bool
	}{
		{
			name: "200",
			h:    HTTPFetcher{},
			args: args{
				uri: "content:foo",
			},
			want:    io.NopCloser(bytes.NewBufferString("foo")),
			wantErr: false,
		},
		{
			name: "404",
			h:    HTTPFetcher{},
			args: args{
				uri: "404",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "401",
			h:    HTTPFetcher{},
			args: args{
				uri: "401",
			},
			want:    nil,
			wantErr: true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("test server download %v must always use GET, used %s instead", r.URL, r.Method)
		}
		// return everything after colon as content to the client, return 200
		if strings.HasPrefix(r.URL.Path, "/content:") {
			if len(r.URL.Path) < 10 {
				w.WriteHeader(http.StatusBadRequest)
				t.Fatalf("test server received content directive with no content")
				return
			}
			content := r.URL.Path[9:]
			_, err := w.Write([]byte(content))
			if err != nil {
				t.Fatalf("test server could not write content response: %s", err)
			}
			return
		}
		// assume that any other resource is just asking to return a specific return code
		i, err := strconv.Atoi(r.URL.Path[1:])
		if err != nil {
			t.Fatalf("test server url received unknown directive: %v", r.URL.Path)
		}
		w.WriteHeader(i)
	}))
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HTTPFetcher{}
			got, err := h.Get(ts.URL + "/" + tt.args.uri)
			if got != nil {
				defer got.Close()
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPFetcher.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (tt.want != nil) && (got == nil) {
				t.Error("HTTPFetcher.Get() = nil, want not nil")
				return
			}
			if (tt.want == nil) && (got != nil) {
				t.Error("HTTPFetcher.Get() != nil, want nil")
				return
			}
			if got == nil && tt.want == nil {
				return
			}
			gotS, err := io.ReadAll(got)
			if err != nil {
				t.Errorf("HTTPFetcher.Get() could not read body: %s", err)
				return
			}
			// this is a local buffer, no error will be returned
			wantS, _ := io.ReadAll(tt.want)
			if !reflect.DeepEqual(gotS, wantS) {
				t.Errorf("HTTPFetcher.Get() = %v, want %v", gotS, wantS)
			}
		})
	}
}
