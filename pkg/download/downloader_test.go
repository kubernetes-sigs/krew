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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func testdataPath() string {
	pwd, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	return filepath.Join(pwd, "testdata")
}

func Test_extract(t *testing.T) {
	// Zip has just one file named 'foo'
	zipSrc := filepath.Join(testdataPath(), "test.zip")
	zipDst, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(zipDst)

	zipReader, err := os.Open(zipSrc)
	if err != nil {
		t.Fatal(err)
	}
	stat, _ := zipReader.Stat()
	if err := extractZIP(zipDst, zipReader, stat.Size()); err != nil {
		t.Fatalf("extract() error = %v", err)
	}
	zipContent, err := ioutil.ReadDir(zipDst)
	if err != nil {
		t.Fatal(err)
	}

	if len(zipContent) != 1 {
		t.Fatalf("zip should just have one file got %d", len(zipContent))
	}
	for _, f := range zipContent {
		if f.IsDir() {
			t.Fatalf("zip should be inflated, got dir %q at root", f.Name())
		}
		if f.Name() != "foo" {
			t.Fatalf("expected to find file foo, found %q", f.Name())
		}
	}
}
