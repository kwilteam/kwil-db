package auth

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func Test_loadKeys(t *testing.T) {
	type args struct {
		h io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]struct{}
		wantErr bool
	}{
		{
			name: "normal",
			args: args{h: strings.NewReader(`{"keys": ["keya", "keyb"]}`)},
			want: map[string]struct{}{
				"keya": {},
				"keyb": {},
			},
			wantErr: false,
		},
		{
			name:    "read empty value",
			args:    args{h: strings.NewReader(`{"keyssss": ["keya", "keyb"]}`)},
			want:    map[string]struct{}{},
			wantErr: false,
		},
		{
			name:    "wrong format should error",
			args:    args{h: strings.NewReader(`{"keys": {"keya": "keyb"}}`)},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadKeys(tt.args.h)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadKeys() got = %v, want %v", got, tt.want)
			}
		})
	}
}
