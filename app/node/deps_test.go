package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
