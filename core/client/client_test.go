package client

import (
	"reflect"
	"testing"
)

func Test_convertTuple(t *testing.T) {
	tests := []struct {
		name      string
		tuple     []any
		want      []string
		wantIsNil []bool
		wantErr   bool
	}{
		{
			"string",
			[]any{"woot"},
			[]string{"woot"},
			[]bool{false},
			false,
		},
		{
			"int",
			[]any{1},
			[]string{"1"},
			[]bool{false},
			false,
		},
		{
			"empty",
			nil,
			[]string{}, // not nil, presently
			[]bool{},
			false,
		},
		{
			"nil",
			[]any{nil},
			[]string{""},
			[]bool{true},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotIsNil, err := convertTuple(tt.tuple)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertTuple() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertTuple() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(gotIsNil, tt.wantIsNil) {
				t.Errorf("convertTuple() gotIsNil = %v, want %v", gotIsNil, tt.wantIsNil)
			}
		})
	}
}
