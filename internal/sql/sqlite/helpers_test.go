package sqlite

import (
	"runtime"
	"testing"
)

func Test_formatFilePathNotWindows(t *testing.T) {
	// This test defines "want" paths in the *NIX path convention used by linux,
	// mac, bsd, etc., but not windows. Rather than using build flags to skip
	// this entire file, we'll define this test as such and we can make another
	// for Windows hosts if we want to.
	if runtime.GOOS == "windows" {
		t.Skip("test not applicable to windows paths")
	}
	type args struct {
		path     string
		fileName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "abs path no slash",
			args: args{
				path:     "/tmp",
				fileName: "dbname",
			},
			want: "/tmp/dbname.sqlite",
		},
		{
			name: "abs path with trailing slash",
			args: args{
				path:     "/tmp/",
				fileName: "dbname",
			},
			want: "/tmp/dbname.sqlite",
		},
		{
			name: "rel path no slash",
			args: args{
				path:     "./here",
				fileName: "dbname",
			},
			want: "here/dbname.sqlite",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatFilePath(tt.args.path, tt.args.fileName); got != tt.want {
				t.Errorf("formatFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
