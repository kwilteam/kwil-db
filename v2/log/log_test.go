package log

import (
	"log/slog"
	"testing"
)

func TestSanitizeArgs(t *testing.T) {
	tests := []struct {
		name string
		args []any
		want []any
	}{
		{
			name: "empty args",
			args: []any{},
			want: []any{},
		},
		{
			name: "single key without value",
			args: []any{"key1"},
			want: []any{"key1"},
		},
		{
			name: "non-string key",
			args: []any{123, "value"},
			want: []any{},
		},
		{
			name: "reserved keys",
			args: []any{
				slog.TimeKey, "time_value",
				slog.LevelKey, "level_value",
				slog.MessageKey, "msg_value",
				slog.SourceKey, "source_value",
			},
			want: []any{
				"timekey", "time_value",
				"levelkey", "level_value",
				"message", "msg_value",
				"soucekey", "source_value",
			},
		},
		{
			name: "mixed valid and invalid keys",
			args: []any{
				"valid1", "value1",
				123, "invalid_value",
				"valid2", "value2",
			},
			want: []any{
				"valid1", "value1",
				"valid2", "value2",
			},
		},
		{
			name: "multiple key-value pairs",
			args: []any{
				"key1", "value1",
				"key2", 42,
				"key3", true,
			},
			want: []any{
				"key1", "value1",
				"key2", 42,
				"key3", true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeArgs(tt.args)
			if len(got) != len(tt.want) {
				t.Errorf("sanitizeArgs() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("sanitizeArgs() index %d = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
