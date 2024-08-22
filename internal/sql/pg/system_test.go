package pg

import (
	"reflect"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"

	"github.com/stretchr/testify/assert"
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

func Test_joinSlices(t *testing.T) {
	var s0 any = []int{}
	// t.Logf("%T: %v", s0, s0) // []int: []

	s1t := []int{1, 2, 3, 4}
	var s1 any = s1t
	s2t := []int{7, 8, 9}
	var s2 any = s2t

	// var s3 []int // nope, s3 is any([]int)
	s3 := joinSlices(s1, s2)

	assert.Equal(t, reflect.TypeOf(s0), reflect.TypeOf(s3))
	t.Log(reflect.TypeOf(s3))

	s3t, ok := s3.([]int)
	if !ok {
		t.Fatalf("not a []int: %T", s3)
	}
	// t.Logf("%T: %v", s3, s3) // []int: [1 2 3 4 7 8 9]

	require.Len(t, s3t, len(s1t)+len(s2t))
	require.EqualValues(t, append(s1t, s2t...), s3t)
}

func TestStatsVal(t *testing.T) {
	tests := []struct {
		name     string
		colType  ColType
		expected any
	}{
		{
			name:     "ColTypeInt",
			colType:  ColTypeInt,
			expected: int64(0),
		},
		{
			name:     "ColTypeText",
			colType:  ColTypeText,
			expected: "",
		},
		{
			name:     "ColTypeBool",
			colType:  ColTypeBool,
			expected: false,
		},
		{
			name:     "ColTypeByteA",
			colType:  ColTypeByteA,
			expected: []byte{},
		},
		{
			name:     "ColTypeUUID",
			colType:  ColTypeUUID,
			expected: &types.UUID{},
		},
		{
			name:     "ColTypeNumeric",
			colType:  ColTypeNumeric,
			expected: &decimal.Decimal{},
		},
		{
			name:     "ColTypeUINT256",
			colType:  ColTypeUINT256,
			expected: &types.Uint256{},
		},
		{
			name:     "ColTypeFloat",
			colType:  ColTypeFloat,
			expected: float64(0),
		},
		{
			name:     "ColTypeTime",
			colType:  ColTypeTime,
			expected: time.Time{},
		},
		{
			name:     "Unknown ColType",
			colType:  ColType("asdfasdf"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statsVal(tt.colType)
			require.IsType(t, tt.expected, result)
			if tt.expected != nil {
				require.Equal(t, reflect.TypeOf(tt.expected), reflect.TypeOf(result))
			}
		})
	}
}
func TestStatsValType(t *testing.T) {
	tests := []struct {
		name     string
		colType  ColType
		expected reflect.Type
	}{
		{
			name:     "ColTypeInt",
			colType:  ColTypeInt,
			expected: reflect.TypeOf(int64(0)),
		},
		{
			name:     "ColTypeText",
			colType:  ColTypeText,
			expected: reflect.TypeOf(""),
		},
		{
			name:     "ColTypeBool",
			colType:  ColTypeBool,
			expected: reflect.TypeOf(false),
		},
		{
			name:     "ColTypeByteA",
			colType:  ColTypeByteA,
			expected: reflect.TypeOf([]byte(nil)),
		},
		{
			name:     "ColTypeUUID",
			colType:  ColTypeUUID,
			expected: reflect.TypeOf(&types.UUID{}),
		},
		{
			name:     "ColTypeNumeric",
			colType:  ColTypeNumeric,
			expected: reflect.TypeOf(&decimal.Decimal{}),
		},
		{
			name:     "ColTypeUINT256",
			colType:  ColTypeUINT256,
			expected: reflect.TypeOf(&types.Uint256{}),
		},
		{
			name:     "ColTypeFloat",
			colType:  ColTypeFloat,
			expected: reflect.TypeOf(float64(0)),
		},
		{
			name:     "ColTypeTime",
			colType:  ColTypeTime,
			expected: reflect.TypeOf(time.Time{}),
		},
		{
			name:     "Unknown ColType",
			colType:  ColType("unknown"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statsValType(tt.colType)
			if tt.expected == nil {
				require.Nil(t, result)
			} else {
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestStatsValTypeConsistency(t *testing.T) {
	allTypes := []ColType{
		ColTypeInt,
		ColTypeText,
		ColTypeBool,
		ColTypeByteA,
		ColTypeUUID,
		ColTypeNumeric,
		ColTypeUINT256,
		ColTypeFloat,
		ColTypeTime,
	}

	for _, ct := range allTypes {
		t.Run(string(ct), func(t *testing.T) {
			valType := statsValType(ct)
			val := statsVal(ct)
			require.NotNil(t, valType)
			require.NotNil(t, val)
			require.Equal(t, valType, reflect.TypeOf(val))
		})
	}
}
