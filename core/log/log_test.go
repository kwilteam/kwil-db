package log

import (
	"bytes"
	"log/slog"
	"strings"
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
				"time_", "time_value",
				"level_", "level_value",
				"msg_", "msg_value",
				"source_", "source_value",
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

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  string
	}{
		{
			name:  "level below debug",
			level: Level(-1),
			want:  "unknown",
		},
		{
			name:  "level above error",
			level: Level(5),
			want:  "unknown",
		},
		{
			name:  "debug",
			level: LevelDebug,
			want:  "debug",
		},
		{
			name:  "info",
			level: LevelInfo,
			want:  "info",
		},
		{
			name:  "warn",
			level: LevelWarn,
			want:  "warn",
		},
		{
			name:  "error",
			level: LevelError,
			want:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("Level.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLevelToSlog(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  slog.Level
	}{
		{
			name:  "negative level",
			level: Level(-1),
			want:  slog.LevelInfo,
		},
		{
			name:  "level above max",
			level: Level(100),
			want:  slog.LevelInfo,
		},
		{
			name:  "debug",
			level: LevelDebug,
			want:  slog.LevelDebug,
		},
		{
			name:  "info",
			level: LevelInfo,
			want:  slog.LevelInfo,
		},
		{
			name:  "warn",
			level: LevelWarn,
			want:  slog.LevelWarn,
		},
		{
			name:  "error",
			level: LevelError,
			want:  slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := levelToSlog(tt.level); got != tt.want {
				t.Errorf("levelToSlog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Level
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid level",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "debug",
			input:   "debug",
			want:    LevelDebug,
			wantErr: false,
		},
		{
			name:    "info",
			input:   "info",
			want:    LevelInfo,
			wantErr: false,
		},
		{
			name:    "warn",
			input:   "warn",
			want:    LevelWarn,
			wantErr: false,
		},
		{
			name:    "error",
			input:   "error",
			want:    LevelError,
			wantErr: false,
		},
		{
			name:    "upper case debug",
			input:   "DEBUG",
			want:    LevelDebug,
			wantErr: false,
		},
		{
			name:    "mixed case info",
			input:   "Info",
			want:    LevelInfo,
			wantErr: false,
		},
		{
			name:    "whitespace",
			input:   "  info  ",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatArgs(t *testing.T) {
	tests := []struct {
		name string
		args []any
		want string
	}{
		{
			name: "empty args",
			args: []any{},
			want: "",
		},
		{
			name: "single key without value",
			args: []any{"key1"},
			want: " {key1}",
		},
		{
			name: "byte slice value",
			args: []any{"bytes", []byte{0xDE, 0xAD, 0xBE, 0xEF}},
			want: " {bytes=deadbeef}",
		},
		{
			name: "multiple key-value pairs with byte slice",
			args: []any{
				"str", "hello",
				"bytes", []byte{0x12, 0x34},
				"num", 42,
			},
			want: " {str=hello bytes=1234 num=42}",
		},
		{
			name: "special characters in keys and values",
			args: []any{
				"key=with=equals", "value",
				"spaces key", "spaces value",
			},
			want: " {key=with=equals=value spaces key=spaces value}",
		},
		{
			name: "nil value",
			args: []any{
				"key1", nil,
				"key2", "value2",
			},
			want: " {key1=<nil> key2=value2}",
		},
		{
			name: "complex types",
			args: []any{
				"array", [3]int{1, 2, 3},
				"slice", []string{"a", "b"},
				"map", map[string]int{"x": 1},
			},
			want: " {array=[1 2 3] slice=[a b] map=map[x:1]}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatArgs(tt.args...); got != tt.want {
				t.Errorf("formatArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlainKVLoggerLogf(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WithWriter(&buf), WithLevel(LevelInfo), WithName("test"))

	tests := []struct {
		level Level
		msg   string
		args  []any
		want  string
	}{
		{LevelDebug, "test message %s", []any{"debug"}, ""}, // below level
		{LevelInfo, "test message %s", []any{"info"}, "test message info"},
		{LevelInfo, "count: %d", []any{42}, "count: 42"},
		{LevelWarn, "warning %s %d", []any{"code", 404}, "warning code 404"},
		{LevelError, "error: %v", []any{"not found"}, "error: not found"},
	}

	for _, tt := range tests {
		buf.Reset()
		logger.Logf(tt.level, tt.msg, tt.args...)
		got := buf.String()
		if tt.want == "" {
			if got != "" {
				t.Errorf("Logf(%v, %q, %v) = %q, want empty string",
					tt.level, tt.msg, tt.args, got)
			}
			continue
		}
		if !strings.Contains(got, tt.want) {
			t.Errorf("Logf(%v, %q, %v) = %q, want string containing %q",
				tt.level, tt.msg, tt.args, got, tt.want)
		}
		if !strings.Contains(got, "system=test") {
			t.Errorf("Logf(%v, %q, %v) = %q, want string containing %q",
				tt.level, tt.msg, tt.args, got, tt.want)
		}
	}
}
