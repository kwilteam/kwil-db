package truflation

import "regexp"

var dateRegexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// The truflation streams protocol understands dates in the format YYYY-MM-DD.
// This function checks if a date is in that format.
// It also recognizes that an empty string is a valid date.
func IsValidDate(date string) bool {
	if date == "" {
		return true
	}
	return dateRegexp.MatchString(date)
}
