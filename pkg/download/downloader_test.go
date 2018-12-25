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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func testdataPath() string {
	pwd, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	return filepath.Join(pwd, "testdata")
}

func Test_extractZIP(t *testing.T) {
	tests := []struct {
		in    string
		files []string
	}{
		{
			in: "test-with-directory.zip",
			files: []string{
				"/test/",
				"/test/foo"}},
		{
			in: "test-without-directory.zip",
			files: []string{
				"/foo"}},
	}

	for _, tt := range tests {
		// Zip has just one file named 'foo'
		zipSrc := filepath.Join(testdataPath(), tt.in)
		zipDst, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(zipDst)

		zipReader, err := os.Open(zipSrc)
		if err != nil {
			t.Fatal(err)
		}
		defer zipReader.Close()
		stat, _ := zipReader.Stat()
		if err := extractZIP(zipDst, zipReader, stat.Size()); err != nil {
			t.Fatalf("extractZIP(%s) error = %v", tt.in, err)
		}

		outFiles := collectFiles(t, zipDst)
		if !reflect.DeepEqual(outFiles, tt.files) {
			t.Fatalf("extractZIP(%s), expected=%#v, got=%#v", tt.in, tt.files, outFiles)
		}
	}
}

func Test_extractTARGZ(t *testing.T) {
	tests := []struct {
		in    string
		files []string
	}{
		{
			in:    "test-without-directory.tar.gz",
			files: []string{"/foo"},
		},
		{
			in: "test-with-nesting-with-directory-entries.tar.gz",
			files: []string{
				"/test/",
				"/test/foo"},
		},
		{
			in: "test-with-nesting-without-directory-entries.tar.gz",
			files: []string{
				"/test/",
				"/test/foo"},
		},
	}

	for _, tt := range tests {
		tarSrc := filepath.Join(testdataPath(), tt.in)
		tarDst, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tarDst)

		tf, err := os.Open(tarSrc)
		if err != nil {
			t.Fatalf("failed to open %q. error=%v", tt.in, err)
		}
		defer tf.Close()
		st, err := tf.Stat()
		if err != nil {
			t.Fatal(err)
			return
		}
		if err := extractTARGZ(tarDst, tf, st.Size()); err != nil {
			t.Fatalf("failed to extract %q. error=%v", tt.in, err)
		}

		outFiles := collectFiles(t, tarDst)
		if !reflect.DeepEqual(outFiles, tt.files) {
			t.Fatalf("for %q, expected=%#v, got=%#v", tt.in, tt.files, outFiles)
		}
	}
}

// collectFiles lists the files by walking the path. It prefixes elements with
// "/" and appends "/" to directories.
func collectFiles(t *testing.T, scanPath string) []string {
	var outFiles []string
	if err := filepath.Walk(scanPath, func(fp string, info os.FileInfo, err error) error {
		if fp == scanPath {
			return nil
		}
		fp = strings.TrimPrefix(fp, scanPath)
		if info.IsDir() {
			fp = fp + "/"
		}
		outFiles = append(outFiles, fp)
		return nil
	}); err != nil {
		t.Fatalf("failed to scan extracted dir %v. error=%v", scanPath, err)
	}
	return outFiles
}

func Test_extractContentType(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "type zip",
			args: args{
				file: filepath.Join(testdataPath(), "test-with-directory.zip"),
			},
			want:    "application/zip",
			wantErr: false,
		},
		{
			name: "type tar.gz",
			args: args{
				file: filepath.Join(testdataPath(), "test-with-nesting-with-directory-entries.tar.gz"),
			},
			want:    "application/x-gzip",
			wantErr: false,
		},
		{
			name: "type elf-bin",
			args: args{
				file: filepath.Join(testdataPath(), "elf-file"),
			},
			want:    "application/octet-stream",
			wantErr: false,
		},
		{
			name: "type null",
			args: args{
				file: filepath.Join(testdataPath(), "null-file"),
			},
			want:    "text/plain; charset=utf-8",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fd, err := os.Open(tt.args.file)
			if err != nil {
				t.Errorf("failed to read file %s, err: %v", tt.args.file, err)
				return
			}
			got, err := extractContentType(fd)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractContentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractArchive(t *testing.T) {
	oldextractors := defaultExtractors
	defer func() {
		defaultExtractors = oldextractors
	}()
	defaultExtractors = map[string]extractor{
		"application/octet-stream":  func(targetDir string, read io.ReaderAt, size int64) error { return nil },
		"text/plain; charset=utf-8": func(targetDir string, read io.ReaderAt, size int64) error { return errors.New("fail test") },
	}
	type args struct {
		filename string
		dst      string
		file     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test extraction",
			args: args{
				filename: "",
				dst:      "",
				file:     filepath.Join(testdataPath(), "elf-file"),
			},
			wantErr: false,
		},
		{
			name: "test fail extraction",
			args: args{
				filename: "",
				dst:      "",
				file:     filepath.Join(testdataPath(), "null-file"),
			},
			wantErr: true,
		},
		{
			name: "test type not found extraction",
			args: args{
				filename: "",
				dst:      "",
				file:     filepath.Join(testdataPath(), "test-with-nesting-with-directory-entries.tar.gz"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fd, err := os.Open(tt.args.file)
			if err != nil {
				t.Errorf("failed to read file %s, err: %v", tt.args.file, err)
				return
			}
			st, err := fd.Stat()
			if err != nil {
				t.Errorf("failed to stat file %s, err: %v", tt.args.file, err)
				return
			}

			if err := extractArchive(tt.args.filename, tt.args.dst, fd, st.Size()); (err != nil) != tt.wantErr {
				t.Errorf("extractArchive() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
