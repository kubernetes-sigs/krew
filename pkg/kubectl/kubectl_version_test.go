package kubectl

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/coreos/go-semver/semver"
)

func mockVersionJSON(ver string) string {
	return fmt.Sprintf(`{"clientVersion": {"gitVersion": %q}}`, ver)
}

func mockVersionCmd(ver string) func() (string, error) {
	return func() (string, error) {
		return mockVersionJSON(ver), nil
	}
}

func badJSONCmd() func() (string, error) {
	return func() (string, error) {
		return `{"clientVersion": {"gitVersion": "v1.11.9"}`, nil
	}
}

func failingVersionCmd() func() (string, error) {
	return func() (string, error) {
		return "", errors.New("mock failure")
	}
}

func Test_parseKubectlVersion(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    *semver.Version
		wantErr bool
	}{
		{name: "valid version with suffix",
			args: mockVersionJSON("v2.0.01"),
			want: semver.New("2.0.1")},
		{name: "valid version with suffix",
			args: mockVersionJSON("v01.11.0-beta.1"),
			want: semver.New("1.11.0-beta.1")},
		{name: "json error",
			args:    `{"clientVersion" {"gitCommit":"1.1.1"}}`,
			wantErr: true},
		{name: "missing field",
			args:    `{"clientVersion": {"compiler":"gc"}}`,
			wantErr: true},
		{name: "invalid version string",
			args:    mockVersionJSON("1.0.a"),
			wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseKubectlVersion(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseKubectlVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseKubectlVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getKubectlVersion(t *testing.T) {
	tests := []struct {
		name        string
		versionExec func() (string, error)
		want        *semver.Version
		wantErr     bool
	}{
		{name: "exec error (mock failure)",
			versionExec: failingVersionCmd(),
			wantErr:     true},
		{name: "parse error (bad json)",
			versionExec: badJSONCmd(),
			wantErr:     true},
		{name: "parse error (bad version)",
			versionExec: mockVersionCmd("v1.a"),
			wantErr:     true},
		{name: "valid",
			versionExec: mockVersionCmd("v1.11.1-alpha.3"),
			want:        semver.New("1.11.1-alpha.3")},
	}

	orig := versionProvider
	defer func() { versionProvider = orig }()

	for _, tt := range tests {
		versionProvider = tt.versionExec
		t.Run(tt.name, func(t *testing.T) {
			got, err := getKubectlVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("getKubectlVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getKubectlVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSupportedVersion(t *testing.T) {
	tests := []struct {
		name            string
		versionExecutor func() (string, error)
		want            bool
		wantErr         bool
	}{
		{name: "valid (equals)",
			versionExecutor: mockVersionCmd("v1.12.0-alpha.0"),
			want:            true},
		{name: "valid (with alpha)",
			versionExecutor: mockVersionCmd("v1.12.0-alpha.0"),
			want:            true},
		{name: "valid (with beta)",
			versionExecutor: mockVersionCmd("v1.12.0-beta.1"),
			want:            true},
		{name: "valid (greater) stable",
			versionExecutor: mockVersionCmd("v1.12.0"),
			want:            true},
		{name: "valid (greater) patch",
			versionExecutor: mockVersionCmd("v1.12.1"),
			want:            true},
		{name: "invalid (unsupported)",
			versionExecutor: mockVersionCmd("v1.11.99999"),
			want:            false},
		{name: "invalid (lexicographically less)",
			versionExecutor: mockVersionCmd("v1.12.0-aa.1"),
			want:            false},
		{name: "exec failure",
			versionExecutor: failingVersionCmd(),
			wantErr:         true}}

	orig := versionProvider
	defer func() { versionProvider = orig }()

	for _, tt := range tests {
		versionProvider = tt.versionExecutor
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsSupportedVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSupportedVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				vv, _ := getKubectlVersion()
				t.Errorf("IsSupportedVersion() = %v (detected=%v) want %v", got, vv, tt.want)
			}
		})
	}
}
