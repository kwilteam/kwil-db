package pg

import "testing"

func Test_validateVersion(t *testing.T) {
	const needMajor, needMinor = 16, 1
	tests := []struct {
		name      string
		pgVerNum  uint32
		wantMajor uint32
		wantMinor uint32
		wantOk    bool
	}{
		{
			"same",
			160001,
			16,
			1,
			true,
		},
		{
			"higher minor, ok",
			160002,
			16,
			2,
			true,
		},
		{
			"lower minor, not ok",
			160000,
			16,
			0,
			false,
		},
		{
			"higher major, not ok",
			170000,
			17,
			0,
			false,
		},
		{
			"lower major, not ok",
			150000,
			15,
			0,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMajor, gotMinor, gotOk := validateVersion(tt.pgVerNum, needMajor, needMinor)
			if gotMajor != tt.wantMajor {
				t.Errorf("validateVersion() gotMajor = %v, want %v", gotMajor, tt.wantMajor)
			}
			if gotMinor != tt.wantMinor {
				t.Errorf("validateVersion() gotMinor = %v, want %v", gotMinor, tt.wantMinor)
			}
			if gotOk != tt.wantOk {
				t.Errorf("validateVersion() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
