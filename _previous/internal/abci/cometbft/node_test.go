package cometbft

import (
	"math"
	"testing"

	"go.uber.org/zap/zapcore"
)

func Test_keyValsToFields(t *testing.T) {
	tests := []struct {
		name           string
		kvs            []any
		wantKeys       []string
		wantValTypes   []zapcore.FieldType
		wantStringVals []string
		wantIntVals    []int64
		wantFloatVals  []float64
	}{
		{
			name:           "two with string key",
			kvs:            []any{"key", "val"},
			wantKeys:       []string{"key"},
			wantValTypes:   []zapcore.FieldType{zapcore.StringType},
			wantStringVals: []string{"val"},
		},
		{
			name:           "missing value",
			kvs:            []any{"key0", "val0", "key1"},
			wantKeys:       []string{"key0", "key1"},
			wantValTypes:   []zapcore.FieldType{zapcore.StringType, zapcore.ErrorType},
			wantStringVals: []string{"val0"}, // last is an error type
		},
		{
			name:           "four with string keys",
			kvs:            []any{"key0", "val0", "key1", "val1"},
			wantKeys:       []string{"key0", "key1"},
			wantValTypes:   []zapcore.FieldType{zapcore.StringType, zapcore.StringType},
			wantStringVals: []string{"val0", "val1"},
		},
		{
			name:           "six with mixed type keys",
			kvs:            []any{"key0", "val0", "key1", "val1", "key2", 222},
			wantKeys:       []string{"key0", "key1", "key2"},
			wantValTypes:   []zapcore.FieldType{zapcore.StringType, zapcore.StringType, zapcore.Int64Type},
			wantStringVals: []string{"val0", "val1", ""},
			wantIntVals:    []int64{0, 0, 222},
		},
		{
			name:         "int64 value",
			kvs:          []any{"key", int64(42)},
			wantKeys:     []string{"key"},
			wantValTypes: []zapcore.FieldType{zapcore.Int64Type},
			wantIntVals:  []int64{42},
		},
		{
			name:         "uint32 value",
			kvs:          []any{"key", uint32(42)},
			wantKeys:     []string{"key"},
			wantValTypes: []zapcore.FieldType{zapcore.Uint32Type},
			wantIntVals:  []int64{42},
		},
		{
			name:          "float64 value",
			kvs:           []any{"key", 42.42},
			wantKeys:      []string{"key"},
			wantValTypes:  []zapcore.FieldType{zapcore.Float64Type},
			wantFloatVals: []float64{42.42},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFields := keyValsToFields(tt.kvs)
			if len(gotFields) != len(tt.wantKeys) {
				t.Fatalf("got %d fields, wanted %d", len(gotFields), len(tt.wantKeys))
			}
			for i, field := range gotFields {
				if tt.wantKeys[i] != field.Key {
					t.Errorf("wanted key %q, got %q", tt.wantKeys[i], field.Key)
				}
				if tt.wantValTypes[i] != field.Type {
					t.Errorf("wanted value type %v, got %v", tt.wantValTypes[i], field.Type)
					continue
				}

				switch field.Type {
				case zapcore.StringType, zapcore.StringerType:
					if field.String != tt.wantStringVals[i] {
						t.Errorf("wanted string value %v, got %v", tt.wantStringVals[i], field.String)
					}
				case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
					if field.Integer != tt.wantIntVals[i] {
						t.Errorf("wanted int value %v, got %v", tt.wantIntVals[i], field.Integer)
					}
				case zapcore.Float64Type, zapcore.Float32Type:
					gotFloat := math.Float64frombits(uint64(field.Integer))
					if gotFloat != tt.wantFloatVals[i] {
						t.Errorf("wanted float value %v, got %v", tt.wantFloatVals[i], gotFloat)
					}
				}
			}
		})
	}
}
