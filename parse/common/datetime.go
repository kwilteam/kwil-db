package common

import (
	"fmt"
	"strings"
	"time"
)

/*
	Kwil's formatting specifiers:
	- YYYY: 4-digit year
	- YY: 2-digit year
	- MM: 2-digit month
	- DD: 2-digit day
	- HH: 2-digit hour (24-hour clock)
	- HH12: 2-digit hour (12-hour clock)
	- MI: 2-digit minute
	- SS: 2-digit second
	- MS: 3-digit millisecond
	- US: 6-digit microsecond
	- A.M.: AM/PM indicator (upper case)
	- a.m.: AM/PM indicator (lower case)
	- P.M.: AM/PM indicator (upper case)
	- p.m.: AM/PM indicator (lower case)
*/

var strftimeReplacer = strings.NewReplacer(
	"YYYY", "2006",
	"YY", "06",
	"MM", "01",
	"DD", "02",
	"HH12", "03",
	"HH", "15",
	"MI", "04",
	"SS", "05",
	"MS", "000",
	"US", "000000",
	"A.M.", "PM",
	"a.m.", "pm",
	"P.M.", "PM",
	"p.m.", "pm",
)

// parseTimestamp converts a timestamp to a microsecond Unix timestamp.
func parseTimestamp(format, value string) (int64, error) {
	layout := strftimeReplacer.Replace(format)

	isPM := false
	if strings.Contains(value, "P.M.") {
		value = strings.ReplaceAll(value, "P.M.", "PM")
		isPM = true
	}
	if strings.Contains(value, "p.m.") {
		value = strings.ReplaceAll(value, "p.m.", "pm")
		isPM = true
	}
	if strings.Contains(value, "A.M.") {
		value = strings.ReplaceAll(value, "A.M.", "AM")
	}
	if strings.Contains(value, "a.m.") {
		value = strings.ReplaceAll(value, "a.m.", "am")
	}

	if !isPM && (strings.Contains(value, "PM") || strings.Contains(value, "pm")) {
		isPM = true
	}

	t, err := time.Parse(layout, value)
	if err != nil {
		return -1, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	if isPM && t.Hour() < 12 {
		t = t.Add(time.Hour * 12)
	}

	return t.UnixMicro(), nil
}

// formatUnixMicro converts a Unix timestamp in microseconds to a formatted string
func formatUnixMicro(unixMicro int64, format string) string {
	t := time.UnixMicro(unixMicro)
	layout := strftimeReplacer.Replace(format)
	return t.UTC().Format(layout)
}
