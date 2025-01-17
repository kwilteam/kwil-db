package cmds

import (
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/csv"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/stretchr/testify/require"
)

func Test_CSVParams(t *testing.T) {
	type testcase struct {
		name         string
		actionParams []string // name and type delimited by :
		csv          *csv.CSV
		mapping      []string
		namedParams  []string
		want         [][]any
		wantErr      bool
	}

	testcases := []testcase{
		{
			name:         "simple, mapping named action param",
			actionParams: []string{"id:int"},
			csv: &csv.CSV{
				Header:  []string{"id"},
				Records: [][]string{{"1"}, {"2"}},
			},
			mapping: []string{"id:id"},
			want:    [][]any{{ptr(int64(1))}, {ptr(int64(2))}},
		},
		{
			name:         "simple, positional parameter from csv",
			actionParams: []string{"id:int"},
			csv: &csv.CSV{
				Header:  []string{"id"},
				Records: [][]string{{"1"}, {"2"}},
			},
			mapping: []string{"id:1"},
			want:    [][]any{{ptr(int64(1))}, {ptr(int64(2))}},
		},
		{
			name:         "csv with 1 val, named param with 1 val",
			actionParams: []string{"id:int", "name:text"},
			csv: &csv.CSV{
				Header:  []string{"id"},
				Records: [][]string{{"1"}, {"2"}},
			},
			mapping:     []string{"id:id"},
			namedParams: []string{"name:text=foo"},
			want:        [][]any{{ptr(int64(1)), ptr("foo")}, {ptr(int64(2)), ptr("foo")}},
		},
		{
			name:         "csv with 1 val, named param (positional) with 1 val",
			actionParams: []string{"id:int", "name:text"},
			csv: &csv.CSV{
				Header:  []string{"id"},
				Records: [][]string{{"1"}, {"2"}},
			},
			mapping:     []string{"id:1"},
			namedParams: []string{"name:text=foo"},
			want:        [][]any{{ptr(int64(1)), ptr("foo")}, {ptr(int64(2)), ptr("foo")}},
		},
		{
			name:         "map to non-existent param",
			actionParams: []string{"id:int"},
			csv: &csv.CSV{
				Header:  []string{"id"},
				Records: [][]string{{"1"}},
			},
			mapping: []string{"id:id2"},
			wantErr: true,
		},
		{
			name:         "map to non-existent csv column",
			actionParams: []string{"id:int"},
			csv: &csv.CSV{
				Header:  []string{"id"},
				Records: [][]string{{"1"}},
			},
			mapping: []string{"id2:id"},
			wantErr: true,
		},
		{
			name:         "non-existent named param",
			actionParams: []string{"id:int"},
			csv: &csv.CSV{
				Header:  []string{"id"},
				Records: [][]string{{"1"}},
			},
			mapping:     []string{"id:id"},
			namedParams: []string{"name:text=foo"},
			wantErr:     true,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var actionParams []NamedParameter
			for _, param := range tt.actionParams {
				parts := strings.Split(param, ":")

				dt, err := types.ParseDataType(parts[1])
				require.NoError(t, err)

				actionParams = append(actionParams, NamedParameter{
					Name: parts[0],
					Type: dt,
				})
			}

			res, err := csvToParams(actionParams, tt.csv, tt.mapping, tt.namedParams)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.EqualValues(t, tt.want, res)
		})
	}
}

func ptr[T any](a T) *T {
	return &a
}
