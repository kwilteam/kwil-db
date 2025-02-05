package node

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ExtensionFlags(t *testing.T) {
	type testcase struct {
		name    string
		flagset []string
		want    map[string]map[string]string
		wantErr bool
	}

	tests := []testcase{
		{
			name:    "empty flagset",
			flagset: []string{},
			want:    map[string]map[string]string{},
		},
		{
			name:    "single flag",
			flagset: []string{"--extensions.extname.flagname", "value"},
			want: map[string]map[string]string{
				"extname": {
					"flagname": "value",
				},
			},
		},
		{
			name:    "multiple flags",
			flagset: []string{"--extensions.extname.flagname", "value", "--extensions.extname2.flagname2=value2"},
			want: map[string]map[string]string{
				"extname": {
					"flagname": "value",
				},
				"extname2": {
					"flagname2": "value2",
				},
			},
		},
		{
			name:    "multiple flags with dot values",
			flagset: []string{"--extensions.extname.flagname", "value.a.b", "--extensions.extname2.flagname2=value2.a.b"},
			want: map[string]map[string]string{
				"extname": {
					"flagname": "value.a.b",
				},
				"extname2": {
					"flagname2": "value2.a.b",
				},
			},
		},
		{
			name:    "more than 3 fields",
			flagset: []string{"--extensions.extname.flagname.another", "value", "--extensions.extname.flagname.another=value"},
			wantErr: true,
		},
		{
			name: "missing value",
			flagset: []string{
				"--extensions.extname.flagname",
			},
			wantErr: true,
		},
		{
			name: "pass flag as a value errors",
			flagset: []string{
				"--extensions.extname.flagname", "--extensions.extname2.flagname2=value2",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExtensionFlags(tt.flagset)
			if tt.wantErr {
				require.Error(t, err)
				t.Log(err)
				return
			}
			require.NoError(t, err)

			require.EqualValues(t, tt.want, got)
		})
	}
}
