package files

import (
	"reflect"
	"testing"
)

// Despite its simplicity, I felt this test was necessary
// in case the path to utils changes

func TestLoadFileFromRoot(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test_load_relative_file",
			args: args{
				path: "keys/test_ethereum.pem",
			},
			want: []byte("4bb214b1f3a0737d758bc3828cdff371e3769fe84a2678da34700cb18d50770e"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadFileFromRoot(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFileFromRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadFileFromRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}
