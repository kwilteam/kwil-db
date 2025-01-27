package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MigrationsMarshalUnmarshal(t *testing.T) {

	tests := []struct {
		name string
		m    *MigrationDeclaration
	}{
		{
			name: "Valid MigrationDeclaration",
			m: &MigrationDeclaration{
				ActivationPeriod: 1000,
				Duration:         14400,
				Timestamp:        "2021-09-01T00:00:00Z",
			},
		},
		{
			name: "Null MigrationDeclaration",
			m:    &MigrationDeclaration{},
		},
		{
			name: "empty timestamp",
			m: &MigrationDeclaration{
				ActivationPeriod: 1000,
				Duration:         14400,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.m.MarshalBinary()
			require.NoError(t, err)

			m2 := &MigrationDeclaration{}
			err = m2.UnmarshalBinary(b)
			require.NoError(t, err)

			require.Equal(t, tt.m, m2)
		})
	}
}
