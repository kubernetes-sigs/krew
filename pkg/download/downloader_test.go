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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"

	"sigs.k8s.io/krew/pkg/testutil"
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
		in        string
		files     []string
		expectErr bool
	}{
		{
			in: "test-with-directory.zip",
			files: []string{
				"/test/",
				"/test/foo",
			},
			expectErr: false,
		},
		{
			in: "test-without-directory.zip",
			files: []string{
				"/foo",
			},
			expectErr: false,
		},
		{
			in: "test-with-symlinks.zip",
			files: []string{
				"/symlinks/",
				"/symlinks/message",
				"/symlinks/msg",
			},
			expectErr: false,
		},
		{
			in:        "test-with-symlinks-escaping-parent.zip",
			files:     nil,
			expectErr: true,
		},
		{
			in: "test-with-symlinks-escaping-parent2.zip",
			files: []string{
				"/escaping-link-test3/",
				"/escaping-link-test3/baz", // this escapes only to the staging area, so expectErr: false
				"/escaping-link-test3/foo/",
				"/escaping-link-test3/foo/bar/",
			},
			expectErr: false,
		},
		{
			in:        "test-with-symlinks-escaping-parent3.tar.gz",
			files:     nil,
			expectErr: true,
		},
		{
			in:        "test-with-symlinks-escaping-absolute.zip",
			files:     nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		// Zip has just one file named 'foo'
		zipSrc := filepath.Join(testdataPath(), tt.in)
		tmpDir, cleanup := testutil.NewTempDir(t)
		defer cleanup()

		zipReader, err := os.Open(zipSrc)
		if err != nil {
			t.Fatal(err)
		}
		defer zipReader.Close()
		stat, _ := zipReader.Stat()
		if err := extractZIP(tmpDir.Root(), zipReader, stat.Size()); err != nil {
			if !tt.expectErr {
				t.Fatalf("extractZIP(%s) error = %v", tt.in, err)
			} else {
				// error was expected, all is good
				continue
			}
		}
		outFiles := collectFiles(t, tmpDir.Root())
		if tt.expectErr {
			t.Fatalf("expected extraction failure %q %v", tt.in, outFiles)
		} else if !reflect.DeepEqual(outFiles, tt.files) {
			t.Fatalf("extractZIP(%s), expected=%v, got=%v", tt.in, tt.files, outFiles)
		}
	}
}

func Test_extractTARGZ(t *testing.T) {
	tests := []struct {
		in        string
		files     []string
		expectErr bool
	}{
		{
			in:        "test-without-directory.tar.gz",
			files:     []string{"/foo"},
			expectErr: false,
		},
		{
			in: "test-with-nesting-with-directory-entries.tar.gz",
			files: []string{
				"/test/",
				"/test/foo",
			},
			expectErr: false,
		},
		{
			in: "test-with-nesting-without-directory-entries.tar.gz",
			files: []string{
				"/test/",
				"/test/foo",
			},
			expectErr: false,
		},
		{
			in: "test-with-symlinks.tar.gz",
			files: []string{
				"/symlinks/",
				"/symlinks/message",
				"/symlinks/msg",
			},
			expectErr: false,
		},
		{
			in:        "test-with-symlinks-escaping-parent.tar.gz",
			files:     nil,
			expectErr: true,
		},
		{
			in: "test-with-symlinks-escaping-parent2.tar.gz",
			files: []string{
				"/escaping-link-test3/",
				"/escaping-link-test3/baz", // this escapes only to the staging area, so expectErr: false
				"/escaping-link-test3/foo/",
				"/escaping-link-test3/foo/bar/",
			},
			expectErr: false,
		},
		{
			in:        "test-with-symlinks-escaping-parent3.tar.gz",
			files:     nil,
			expectErr: true,
		},
		{
			in:        "test-with-symlinks-escaping-absolute.tar.gz",
			files:     nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		tarSrc := filepath.Join(testdataPath(), tt.in)
		tmpDir, cleanup := testutil.NewTempDir(t)
		defer cleanup()

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
		if err = extractTARGZ(tmpDir.Root(), tf, st.Size()); err != nil {
			if !tt.expectErr {
				t.Fatalf("failed to extract %q. error=%v", tt.in, err)
			} else {
				// error was expected, all is good
				continue
			}
		}
		outFiles := collectFiles(t, tmpDir.Root())
		if tt.expectErr {
			t.Fatalf("expected extraction failure %q %v", tt.in, outFiles)
		} else if !reflect.DeepEqual(outFiles, tt.files) {
			t.Fatalf("for %q, expected=%v, got=%v", tt.in, tt.files, outFiles)
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
			fp += "/"
		}
		outFiles = append(outFiles, fp)
		return nil
	}); err != nil {
		t.Fatalf("failed to scan extracted dir %v. error=%v", scanPath, err)
	}
	return outFiles
}

func TestDownloader_Get(t *testing.T) {
	type fields struct {
		verifier Verifier
		fetcher  Fetcher
	}
	tests := []struct {
		name    string
		fields  fields
		uri     string
		wantErr bool
	}{
		{
			name: "successful get",
			fields: fields{
				verifier: NewInsecureVerifier(),
				fetcher:  NewFileFetcher(filepath.Join(testdataPath(), "test-with-directory.zip")),
			},
			uri:     "foo/bar/test-with-directory.zip",
			wantErr: false,
		},
		{
			name: "fail get by fetching",
			fields: fields{
				verifier: NewInsecureVerifier(),
				fetcher:  errorFetcher{},
			},
			uri:     "foo/bar/test-with-directory.zip",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.NewTempDir(t)
			defer cleanup()

			d := NewDownloader(tt.fields.verifier, tt.fields.fetcher)
			if err := d.Get(tt.uri, tmpDir.Root()); (err != nil) != tt.wantErr {
				t.Errorf("Downloader.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_download(t *testing.T) {
	filePath := filepath.Join(testdataPath(), "test-with-directory.zip")
	downloadOriginal, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
		return
	}
	type args struct {
		url      string
		verifier Verifier
		fetcher  Fetcher
	}
	tests := []struct {
		name       string
		args       args
		wantReader io.ReaderAt
		wantSize   int64
		wantErr    bool
	}{
		{
			name: "successful fetch",
			args: args{
				url:      filePath,
				verifier: NewInsecureVerifier(),
				fetcher:  NewFileFetcher(filePath),
			},
			wantReader: bytes.NewReader(downloadOriginal),
			wantSize:   int64(len(downloadOriginal)),
			wantErr:    false,
		},
		{
			name: "wrong data fetch",
			args: args{
				url:      filePath,
				verifier: newFalseVerifier(),
				fetcher:  NewFileFetcher(filePath),
			},
			wantReader: bytes.NewReader(downloadOriginal),
			wantSize:   int64(len(downloadOriginal)),
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, size, err := download(tt.args.url, tt.args.verifier, tt.args.fetcher)
			if (err != nil) != tt.wantErr {
				t.Errorf("download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			downloadedData, err := ioutil.ReadAll(io.NewSectionReader(reader, 0, size))
			if err != nil {
				t.Errorf("failed to read download data: %v", err)
				return
			}
			wantData, err := ioutil.ReadAll(io.NewSectionReader(tt.wantReader, 0, tt.wantSize))
			if err != nil {
				t.Errorf("failed to read download data: %v", err)
				return
			}

			if !bytes.Equal(downloadedData, wantData) {
				t.Errorf("download() reader = %v, wantReader %v", reader, tt.wantReader)
			}
			if size != tt.wantSize {
				t.Errorf("download() size = %v, wantReader %v", size, tt.wantSize)
			}
		})
	}
}

var _ Verifier = falseVerifier{}

type falseVerifier struct{ io.Writer }

func newFalseVerifier() Verifier    { return falseVerifier{ioutil.Discard} }
func (falseVerifier) Verify() error { return errors.New("test verifier") }

func Test_detectMIMEType(t *testing.T) {
	type args struct {
		file    string
		content []byte
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
			name: "type bash-utf8",
			args: args{
				file: filepath.Join(testdataPath(), "bash-utf8-file"),
			},
			want:    "text/plain",
			wantErr: false,
		},

		{
			name: "type bash-ascii",
			args: args{
				file: filepath.Join(testdataPath(), "bash-ascii-file"),
			},
			want:    "text/plain",
			wantErr: false,
		},
		{
			name: "type null",
			args: args{
				file: filepath.Join(testdataPath(), "null-file"),
			},
			want:    "text/plain",
			wantErr: false,
		},
		{
			name: "512 zero bytes",
			args: args{
				content: make([]byte, 512),
			},
			want:    "application/octet-stream",
			wantErr: false,
		},
		{
			name: "1 zero bytes",
			args: args{
				content: make([]byte, 1),
			},
			want:    "application/octet-stream",
			wantErr: false,
		},
		{
			name: "0 zero bytes",
			args: args{
				content: []byte{},
			},
			want:    "text/plain",
			wantErr: false,
		},
		{
			name: "html",
			args: args{
				content: []byte("<!DOCTYPE html><html><head><title></title></head><body></body></html>"),
			},
			want:    "text/html",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var at io.ReaderAt

			if tt.args.file != "" {
				fd, err := os.Open(tt.args.file)
				if err != nil {
					t.Errorf("failed to read file %s, err: %v", tt.args.file, err)
					return
				}
				defer fd.Close()
				at = fd
			} else {
				at = bytes.NewReader(tt.args.content)
			}

			got, err := detectMIMEType(at)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectMIMEType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("detectMIMEType() = %v, want %v", got, tt.want)
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
		"application/octet-stream": func(targetDir string, read io.ReaderAt, size int64) error { return nil },
		"text/plain":               func(targetDir string, read io.ReaderAt, size int64) error { return errors.New("fail test") },
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

			if err := extractArchive(tt.args.dst, fd, st.Size()); (err != nil) != tt.wantErr {
				t.Errorf("extractArchive() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
