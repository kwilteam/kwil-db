package modules

import (
	"reflect"
	"testing"
)

func TestConvertModuleError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want *ABCIModuleError
	}{
		{
			name: "value",
			err: ABCIModuleError{
				Code:   1,
				Detail: "a",
			},
			want: &ABCIModuleError{
				Code:   1,
				Detail: "a",
			},
		},
		{
			name: "pointer",
			err: &ABCIModuleError{
				Code:   1,
				Detail: "a",
			},
			want: &ABCIModuleError{
				Code:   1,
				Detail: "a",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertModuleError(tt.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertModuleError() = %v, want %v", got, tt.want)
			}
		})
	}
}
