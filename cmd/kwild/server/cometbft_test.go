package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_cleanListenAddr(t *testing.T) {
	type args struct {
		addr        string
		defaultPort string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"ok no change",
			args{
				addr:        "tcp://127.0.0.1:9090",
				defaultPort: "8080",
			},
			"tcp://127.0.0.1:9090",
		},
		{
			"no port or scheme",
			args{
				addr:        "127.0.0.1",
				defaultPort: "8080",
			},
			"tcp://127.0.0.1:8080",
		},
		{
			"no scheme",
			args{
				addr:        "127.0.0.1:9090",
				defaultPort: "8080",
			},
			"tcp://127.0.0.1:9090",
		},
		{
			"ok no change",
			args{
				addr:        "tcp://localhost:9090",
				defaultPort: "8080",
			},
			"tcp://localhost:9090",
		},
		{
			"no port or scheme",
			args{
				addr:        "localhost",
				defaultPort: "8080",
			},
			"tcp://localhost:8080",
		},
		{
			"no scheme",
			args{
				addr:        "localhost:9090",
				defaultPort: "8080",
			},
			"tcp://localhost:9090",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanListenAddr(tt.args.addr, tt.args.defaultPort); got != tt.want {
				t.Errorf("cleanListenAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPGVersion(t *testing.T) {
	tests := []struct {
		name           string
		versionOutput  string
		expectedMajor  int
		expectedMinor  int
		expectedErrMsg string
	}{
		{
			name:          "Valid version string",
			versionOutput: "psql (PostgreSQL) 14.5",
			expectedMajor: 14,
			expectedMinor: 5,
		},
		{
			name:          "Valid version string with patch",
			versionOutput: "psql (PostgreSQL) 13.2.1",
			expectedMajor: 13,
			expectedMinor: 2,
		},
		{
			name:           "Invalid version string",
			versionOutput:  "psql (PostgreSQL) invalid",
			expectedErrMsg: "could not find a valid version in output: psql (PostgreSQL) invalid",
		},
		{
			name:           "Empty version string",
			versionOutput:  "",
			expectedErrMsg: "could not find a valid version in output: ",
		},
		{
			name:          "Version with extra information",
			versionOutput: "psql (PostgreSQL) 15.3 (Debian 15.3-1.pgdg110+1)",
			expectedMajor: 15,
			expectedMinor: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, err := getPGVersion(tt.versionOutput)

			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMajor, major)
				assert.Equal(t, tt.expectedMinor, minor)
			}
		})
	}
}
