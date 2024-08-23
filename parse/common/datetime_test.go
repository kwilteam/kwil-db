package common

import (
	"testing"
	"time"
)

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		value    string
		expected int64
	}{
		{
			name:     "Basic date",
			format:   "YYYY-MM-DD",
			value:    "2023-05-15",
			expected: time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC).UnixMicro(),
		},
		{
			name:     "Date and time",
			format:   "YYYY-MM-DD HH:MI:SS",
			value:    "2023-05-15 14:30:45",
			expected: time.Date(2023, 5, 15, 14, 30, 45, 0, time.UTC).UnixMicro(),
		},
		{
			name:     "Date and time with microseconds",
			format:   "YYYY-MM-DD HH:MI:SS.US",
			value:    "2023-05-15 14:30:45.123456",
			expected: time.Date(2023, 5, 15, 14, 30, 45, 123456000, time.UTC).UnixMicro(),
		},
		{
			name:     "12-hour clock PM",
			format:   "YYYY-MM-DD HH12:MI:SS A.M.",
			value:    "2023-05-15 02:30:45 P.M.",
			expected: time.Date(2023, 5, 15, 14, 30, 45, 0, time.UTC).UnixMicro(),
		},
		{
			name:     "12-hour clock AM",
			format:   "YYYY-MM-DD HH12:MI:SS a.m.",
			value:    "2023-05-15 10:30:45 a.m.",
			expected: time.Date(2023, 5, 15, 10, 30, 45, 0, time.UTC).UnixMicro(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimestamp(tt.format, tt.value)
			if err != nil {
				t.Errorf("parseTimestamp() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("parseTimestamp() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatUnixMicro(t *testing.T) {
	tests := []struct {
		name      string
		unixMicro int64
		format    string
		expected  string
	}{
		{
			name:      "Basic date",
			unixMicro: time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC).UnixMicro(),
			format:    "YYYY-MM-DD",
			expected:  "2023-05-15",
		},
		{
			name:      "Date and time",
			unixMicro: time.Date(2023, 5, 15, 14, 30, 45, 0, time.UTC).UnixMicro(),
			format:    "YYYY-MM-DD HH:MI:SS",
			expected:  "2023-05-15 14:30:45",
		},
		{
			name:      "Date and time with microseconds",
			unixMicro: time.Date(2023, 5, 15, 14, 30, 45, 123456000, time.UTC).UnixMicro(),
			format:    "YYYY-MM-DD HH:MI:SS.US",
			expected:  "2023-05-15 14:30:45.123456",
		},
		{
			name:      "12-hour clock PM",
			unixMicro: time.Date(2023, 5, 15, 14, 30, 45, 0, time.UTC).UnixMicro(),
			format:    "YYYY-MM-DD HH12:MI:SS P.M.",
			expected:  "2023-05-15 02:30:45 PM",
		},
		{
			name:      "12-hour clock AM",
			unixMicro: time.Date(2023, 5, 15, 10, 30, 45, 0, time.UTC).UnixMicro(),
			format:    "YYYY-MM-DD HH12:MI:SS a.m.",
			expected:  "2023-05-15 10:30:45 am",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUnixMicro(tt.unixMicro, tt.format)
			if result != tt.expected {
				t.Errorf("formatUnixMicro() = %v, want %v", result, tt.expected)
			}
		})
	}
}
