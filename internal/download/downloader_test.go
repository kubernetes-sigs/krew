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
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"

	"sigs.k8s.io/krew/internal/testutil"
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
			in: "test-flat-hierarchy.zip",
			files: []string{
				"/foo",
			},
		},
		{
			in: "test-with-directory-entry.zip",
			files: []string{
				"/test/",
				"/test/foo",
			},
		},
		{
			in: "test-with-no-directory-entry.zip",
			files: []string{
				"/test/",
				"/test/foo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			// Zip has just one file named 'foo'
			zipSrc := filepath.Join(testdataPath(), tt.in)
			tmpDir := testutil.NewTempDir(t)

			zipReader, err := os.Open(zipSrc)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { zipReader.Close() })
			stat, _ := zipReader.Stat()
			if err := extractZIP(tmpDir.Root(), zipReader, stat.Size()); err != nil {
				t.Fatalf("extractZIP(%s) error = %v", tt.in, err)
			}

			outFiles := collectFiles(t, tmpDir.Root())
			if !reflect.DeepEqual(outFiles, tt.files) {
				t.Fatalf("extractZIP(%s), expected=%v, got=%v", tt.in, tt.files, outFiles)
			}
		})
	}
}

func Test_extractTARGZ(t *testing.T) {
	tests := []struct {
		in    string
		files []string
	}{
		{
			in:    "test-flat-hierarchy.tar.gz",
			files: []string{"/foo"},
		},
		{
			in: "test-with-directory-entry.tar.gz",
			files: []string{
				"/test/",
				"/test/foo",
			},
		},
		{
			in: "test-with-no-directory-entry.tar.gz",
			files: []string{
				"/test/",
				"/test/foo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			tarSrc := filepath.Join(testdataPath(), tt.in)
			tmpDir := testutil.NewTempDir(t)

			tf, err := os.Open(tarSrc)
			if err != nil {
				t.Fatalf("failed to open %q. error=%v", tt.in, err)
			}
			t.Cleanup(func() { tf.Close() })
			st, err := tf.Stat()
			if err != nil {
				t.Fatal(err)
				return
			}
			if err := extractTARGZ(tmpDir.Root(), tf, st.Size()); err != nil {
				t.Fatalf("failed to extract %q. error=%v", tt.in, err)
			}

			outFiles := collectFiles(t, tmpDir.Root())
			if !reflect.DeepEqual(outFiles, tt.files) {
				t.Fatalf("for %q, expected=%v, got=%v", tt.in, tt.files, outFiles)
			}
		})
	}
}

// collectFiles lists the files by walking the path. It prefixes elements with
// "/" and appends "/" to directories.
func collectFiles(t *testing.T, scanPath string) []string {
	var outFiles []string
	if err := filepath.Walk(scanPath, func(fp string, info os.FileInfo, _ error) error {
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

var _ Fetcher = errorFetcher{}

type errorFetcher struct{}

func (f errorFetcher) Get(_ string) (io.ReadCloser, error) { return nil, errors.New("test fail") }

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
				verifier: newTrueVerifier(),
				fetcher:  NewFileFetcher(filepath.Join(testdataPath(), "test-with-directory-entry.zip")),
			},
			uri:     "foo/bar/test-with-directory-entry.zip",
			wantErr: false,
		},
		{
			name: "fail get by fetching",
			fields: fields{
				verifier: newTrueVerifier(),
				fetcher:  errorFetcher{},
			},
			uri:     "foo/bar/test-with-directory-entry.zip",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := testutil.NewTempDir(t)

			d := NewDownloader(tt.fields.verifier, tt.fields.fetcher)
			if err := d.Get(tt.uri, tmpDir.Root()); (err != nil) != tt.wantErr {
				t.Errorf("Downloader.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_download(t *testing.T) {
	filePath := filepath.Join(testdataPath(), "test-with-directory-entry.zip")
	downloadOriginal, err := os.ReadFile(filePath)
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
				verifier: newTrueVerifier(),
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
			downloadedData, err := io.ReadAll(io.NewSectionReader(reader, 0, size))
			if err != nil {
				t.Errorf("failed to read download data: %v", err)
				return
			}
			wantData, err := io.ReadAll(io.NewSectionReader(tt.wantReader, 0, tt.wantSize))
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

var _ Verifier = trueVerifier{}

type trueVerifier struct{ io.Writer }

// newTrueVerifier returns a Verifier that always verifies to true.
func newTrueVerifier() Verifier    { return trueVerifier{io.Discard} }
func (trueVerifier) Verify() error { return nil }

var _ Verifier = falseVerifier{}

type falseVerifier struct{ io.Writer }

func newFalseVerifier() Verifier    { return falseVerifier{io.Discard} }
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
				file: filepath.Join(testdataPath(), "test-with-directory-entry.zip"),
			},
			want:    "application/zip",
			wantErr: false,
		},
		{
			name: "type tar.gz",
			args: args{
				file: filepath.Join(testdataPath(), "test-with-directory-entry.tar.gz"),
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
		"application/octet-stream": func(_ string, _ io.ReaderAt, _ int64) error { return nil },
		"text/plain":               func(_ string, _ io.ReaderAt, _ int64) error { return errors.New("fail test") },
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
				file:     filepath.Join(testdataPath(), "test-with-directory-entry.tar.gz"),
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

func Test_suspiciousPath(t *testing.T) {
	tests := []struct {
		path      string
		shouldErr bool
	}{
		{
			path:      `/foo`,
			shouldErr: true,
		},
		{
			path:      `\foo`,
			shouldErr: true,
		},
		{
			path:      `//foo`,
			shouldErr: true,
		},
		{
			path:      `/\foo`,
			shouldErr: true,
		},
		{
			path:      `\\foo`,
			shouldErr: true,
		},
		{
			path: `./foo`,
		},
		{
			path: `././foo`,
		},
		{
			path: `.//foo`,
		},
		{
			path:      `../foo`,
			shouldErr: true,
		},
		{
			path:      `a/../foo`,
			shouldErr: true,
		},
		{
			path: `a/././foo`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			err := suspiciousPath(tt.path)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected suspiciousPath to fail")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected suspiciousPath not to fail, got %s", err)
			}
		})
	}
}

func Test_extractMaliciousArchive(t *testing.T) {
	const testContent = "some file content"

	tests := []struct {
		name string
		path string
	}{
		{
			name: "absolute file",
			path: "/foo",
		},
		{
			name: "contains ..",
			path: "a/../foo",
		},
	}

	for _, tt := range tests {
		t.Run("tar.gz  "+tt.name, func(t *testing.T) {
			tmpDir := testutil.NewTempDir(t)

			// do not use filepath.Join here, because it calls filepath.Clean on the result
			reader, err := tarGZArchiveForTesting(map[string]string{tt.path: testContent})
			if err != nil {
				t.Fatal(err)
			}

			err = extractTARGZ(tmpDir.Root(), reader, reader.Size())
			if err == nil {
				t.Errorf("Expected extractTARGZ to fail")
			} else if !strings.HasPrefix(err.Error(), "refusing to unpack archive") {
				t.Errorf("Found the wrong error: %s", err)
			}
		})
	}

	for _, tt := range tests {
		t.Run("zip  "+tt.name, func(t *testing.T) {
			tmpDir := testutil.NewTempDir(t)

			// do not use filepath.Join here, because it calls filepath.Clean on the result
			reader, err := zipArchiveReaderForTesting(map[string]string{tt.path: testContent})
			if err != nil {
				t.Fatal(err)
			}

			err = extractZIP(tmpDir.Root(), reader, reader.Size())
			if err == nil {
				t.Errorf("Expected extractZIP to fail")
			} else if !strings.HasPrefix(err.Error(), "refusing to unpack archive") {
				t.Errorf("Found the wrong error: %s", err)
			}
		})
	}
}

// tarGZArchiveForTesting creates an in-memory zip archive with entries from
// the files map, where keys are the paths and values are the contents.
// For example, to create an empty file `a` and another file `b/c`:
//
//	tarGZArchiveForTesting(map[string]string{
//	   "a": "",
//	   "b/c": "nested content",
//	})
func tarGZArchiveForTesting(files map[string]string) (*bytes.Reader, error) {
	archiveBuffer := &bytes.Buffer{}
	gzArchiveBuffer := gzip.NewWriter(archiveBuffer)
	tw := tar.NewWriter(gzArchiveBuffer)
	for path, content := range files {
		header := &tar.Header{
			Name: path,
			Size: int64(len(content)),
			Mode: 0o600,
		}
		if err := tw.WriteHeader(header); err != nil {
			return nil, err
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			return nil, err
		}

	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gzArchiveBuffer.Close(); err != nil {
		return nil, err
	}
	return bytes.NewReader(archiveBuffer.Bytes()), nil
}

// zipArchiveReaderForTesting creates an in-memory zip archive with entries from
// the files map, where keys are the paths and values are the contents. Note that
// entries with empty content just create a directory. The zip spec requires that
// parent directories are explicitly listed in the archive, so this must be done
// for nested entries. For example, to create a file at `a/b/c`, you must pass:
//
//	map[string]string{"a": "", "a/b": "", "a/b/c": "nested content"}
func zipArchiveReaderForTesting(files map[string]string) (*bytes.Reader, error) {
	archiveBuffer := &bytes.Buffer{}
	zw := zip.NewWriter(archiveBuffer)
	for path, content := range files {
		f, err := zw.Create(path)
		if err != nil {
			return nil, err
		}
		if content == "" {
			continue
		}
		if _, err := f.Write([]byte(content)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return bytes.NewReader(archiveBuffer.Bytes()), nil
}
