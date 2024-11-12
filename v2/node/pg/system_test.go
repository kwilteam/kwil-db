package pg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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

func Test_wantStringFn(t *testing.T) {
	tests := []struct {
		name    string
		fn      settingValidFn
		check   string
		wantErr bool
	}{
		{
			"ok equal",
			wantStringFn("beef"),
			"beef",
			false,
		},
		{
			"ok equal",
			wantStringFn("beef"),
			"chicken",
			true,
		},
		{
			"ok no case",
			wantStringFn("beef"),
			"BEEf",
			false,
		},
		{
			"ok space prefix suffix on both",
			wantStringFn(" beef"),
			"  beef ",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.check)
			if gotErr := err != nil; tt.wantErr != gotErr {
				t.Errorf("want err %v, got %v", tt.wantErr, gotErr)
			}
		})
	}
}

func Test_wantOnFn(t *testing.T) {
	tests := []struct {
		name    string
		fn      settingValidFn
		check   string
		wantErr bool
	}{
		{
			"ok on",
			wantOnFn(true),
			"on",
			false,
		},
		{
			"ok off",
			wantOnFn(false),
			"off",
			false,
		},
		{
			"not ok on",
			wantOnFn(false),
			"on",
			true,
		},
		{
			"not ok off",
			wantOnFn(true),
			"off",
			true,
		},
		{
			"ok no case or space",
			wantOnFn(true),
			" On	",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.check)
			if gotErr := err != nil; tt.wantErr != gotErr {
				t.Errorf("want err %v, got %v", tt.wantErr, gotErr)
			}
		})
	}
}

func Test_wantMinIntFn(t *testing.T) {
	tests := []struct {
		name    string
		fn      settingValidFn
		check   string
		wantErr bool
	}{
		{
			"ok equal",
			wantMinIntFn(1),
			"1",
			false,
		},
		{
			"ok more",
			wantMinIntFn(1),
			"2",
			false,
		},
		{
			"not ok less",
			wantMinIntFn(1),
			"0",
			true,
		},
		{
			"ok negative want and val",
			wantMinIntFn(-2),
			"-1",
			false,
		},
		{
			"not ok negative val",
			wantMinIntFn(1),
			"-1",
			true,
		},
		{
			"ok negative want",
			wantMinIntFn(-1),
			"1",
			false,
		},
		{
			"float not an integer",
			wantMinIntFn(1),
			"2.2",
			true,
		},
		{
			"text not an integer",
			wantMinIntFn(1),
			"nope",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.check)
			if gotErr := err != nil; tt.wantErr != gotErr {
				t.Errorf("want err %v, got %v", tt.wantErr, gotErr)
			}
		})
	}
}

func TestSettingValidFnAND(t *testing.T) {
	tests := []struct {
		name    string
		fns     []settingValidFn
		val     string
		wantErr bool
	}{
		{
			name:    "no conditions",
			fns:     []settingValidFn{},
			val:     "any",
			wantErr: false,
		},
		{
			name:    "one condition satisfied",
			fns:     []settingValidFn{wantStringFn("abc")},
			val:     "abc",
			wantErr: false,
		},
		{
			name:    "one condition not satisfied",
			fns:     []settingValidFn{wantStringFn("abc")},
			val:     "xyz",
			wantErr: true,
		},
		{
			name:    "multiple conditions, one satisfied",
			fns:     []settingValidFn{wantStringFn("abc"), wantStringFn("xyz")},
			val:     "xyz",
			wantErr: true,
		},
		{
			name:    "multiple conditions, none satisfied",
			fns:     []settingValidFn{wantStringFn("invalid"), wantOnFn(false)},
			val:     "on",
			wantErr: true,
		},
		{
			name:    "multiple int conditions, all satisfied",
			fns:     []settingValidFn{wantMinIntFn(20), wantMaxIntFn(30)},
			val:     "25",
			wantErr: false,
		},
		{
			name:    "multiple int conditions, first satisfied",
			fns:     []settingValidFn{wantMinIntFn(20), wantMaxIntFn(30)},
			val:     "31",
			wantErr: true,
		},
		{
			name:    "multiple int conditions, last satisfied",
			fns:     []settingValidFn{wantMinIntFn(20), wantMaxIntFn(30)},
			val:     "19",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := andValidFn(tt.fns...)
			err := fn(tt.val)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSettingValidFnOR(t *testing.T) {
	tests := []struct {
		name    string
		fns     []settingValidFn
		val     string
		wantErr bool
	}{
		{
			name:    "no conditions",
			fns:     []settingValidFn{},
			val:     "any",
			wantErr: true,
		},
		{
			name:    "one condition satisfied",
			fns:     []settingValidFn{wantStringFn("abc")},
			val:     "abc",
			wantErr: false,
		},
		{
			name:    "one condition not satisfied",
			fns:     []settingValidFn{wantStringFn("abc")},
			val:     "xyz",
			wantErr: true,
		},
		{
			name:    "multiple conditions, one satisfied",
			fns:     []settingValidFn{wantStringFn("abc"), wantStringFn("xyz")},
			val:     "xyz",
			wantErr: false,
		},
		{
			name:    "multiple conditions, none satisfied",
			fns:     []settingValidFn{wantStringFn("invalid"), wantOnFn(false)},
			val:     "on",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := orValidFn(tt.fns...)
			err := fn(tt.val)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
