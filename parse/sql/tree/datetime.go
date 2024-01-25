package tree

import (
	"fmt"
	"strconv"
	"strings"

	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

/*
SQLite DateTime functions are really weird.  Would recommend reading the docs for them here: https://www.sqlite.org/lang_datefunc.html

Here, I outline exactly what we can and cannot support.

Essentially, If they take no args, they will use the machine's local time, which should not be allowed in our case.
This is with the exception of STRFTIME, which will use the local time unless the second argument is specified.
Ex:

	date(time-value, modifier, modifier, ...)
	time(time-value, modifier, modifier, ...)
	datetime(time-value, modifier, modifier, ...)
	julianday(time-value, modifier, modifier, ...)
	unixepoch(time-value, modifier, modifier, ...)
	strftime(format, time-value, modifier, modifier, ...)

Therefore, we need to make sure that the first argument is always specified, and that the second argument is specified for STRFTIME.
We also cannot allow julianday, since it will return a float, which we cannot guarantee precision for.

There are also some restrictions on the string formatting substitutions.  Here are all supported by SQLite:

	%d		day of month: 00
	%f		fractional seconds: SS.SSS
	%H		hour: 00-24
	%j		day of year: 001-366
	%J		Julian day number (fractional)
	%m		month: 01-12
	%M		minute: 00-59
	%s		seconds since 1970-01-01
	%S		seconds: 00-59
	%w		day of week 0-6 with Sunday==0
	%W		week of year: 00-53
	%Y		year: 0000-9999
	%%		%

We cannot allow the following:

	%f		uses fractional seconds, which we cannot guarantee precision for
	%J		uses Julian day number, which we cannot guarantee precision for
	%s		uses seconds since 1970-01-01, which we relies on the machine's local time

We also have to guard against certain time strings. Here are all supported by SQLite:

	YYYY-MM-DD
	YYYY-MM-DD HH:MM
	YYYY-MM-DD HH:MM:SS
	YYYY-MM-DD HH:MM:SS.SSS
	YYYY-MM-DDTHH:MM
	YYYY-MM-DDTHH:MM:SS
	YYYY-MM-DDTHH:MM:SS.SSS
	HH:MM
	HH:MM:SS
	HH:MM:SS.SSS
	now
	DDDDDDDDDD

We cannot allow the following:

	YYYY-MM-DD HH:MM:SS.SSS 	uses fractional seconds, which we cannot guarantee precision for
	YYYY-MM-DDTHH:MM:SS.SSS 	uses fractional seconds, which we cannot guarantee precision for
	HH:MM:SS.SSS 				uses fractional seconds, which we cannot guarantee precision for
	now 						uses the machine's local time
	DDDDDDDDDD 					uses floating point, which we cannot guarantee precision for

Finally, we need to be concerned with modifiers.  Below is a list of all modifiers:

	NNN days
	NNN hours
	NNN minutes
	NNN.NNNN seconds
	NNN months
	NNN years
	start of month
	start of year
	start of day
	weekday N
	unixepoch
	julianday
	auto
	localtime
	utc

We cannot allow the following:

	NNN.NNNN seconds 	uses fractional seconds, which we cannot guarantee precision for
	julianday 			uses floating point, which we cannot guarantee precision for
	auto 				will auto-detect either unixepoch or julianday, and julianday is not supported
	localtime 			uses the machine's local time
	utc 				uses the machine's local time
*/
var (
	FunctionSTRFTIME = DateTimeFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "strftime",
			Min:          2,
		}}
	FunctionDATE = DateTimeFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "date",
			Min:          1,
		}}
	FunctionTIME = DateTimeFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "time",
			Min:          1,
		}}
	FunctionDATETIME = DateTimeFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "datetime",
			Min:          1,
		}}
	FunctionUNIXEPOCH = DateTimeFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "unixepoch",
			Min:          1,
		}}
)

type DateTimeFunction struct {
	node

	AnySQLFunction
}

func (d *DateTimeFunction) Accept(v AstVisitor) any {
	return v.VisitDateTimeFunc(d)
}

func (d *DateTimeFunction) Walk(w AstWalker) error {
	return run(
		w.EnterDateTimeFunc(d),
		w.ExitDateTimeFunc(d),
	)
}

func NewDateTimeFunctionWithGetter(name string, min uint8, max uint8, distinct bool) SQLFunctionGetter {
	return func(pos *Position) SQLFunction {
		return &DateTimeFunction{
			AnySQLFunction: AnySQLFunction{
				FunctionName: name,
				Min:          min,
				Max:          max,
				distinct:     distinct,
			},
		}
	}
}

var (
	FunctionSTRFTIMEGetter  = NewDateTimeFunctionWithGetter("strftime", 2, 0, false)
	FunctionDATEGetter      = NewDateTimeFunctionWithGetter("date", 1, 0, false)
	FunctionTIMEGetter      = NewDateTimeFunctionWithGetter("time", 1, 0, false)
	FunctionDATETIMEGetter  = NewDateTimeFunctionWithGetter("datetime", 1, 0, false)
	FunctionUNIXEPOCHGetter = NewDateTimeFunctionWithGetter("unixepoch", 1, 0, false)
)

func (d *DateTimeFunction) ToString(exprs ...Expression) string {
	if len(exprs) < int(d.Min) {
		panic("not enough arguments for datetime function '" + d.FunctionName + "'")
	}

	if len(exprs) > int(d.Max) && d.Max > 0 {
		panic("too many arguments for datetime function '" + d.FunctionName + "'")
	}

	if d.FunctionName == "strftime" {
		return d.stringStrftime(exprs)
	}

	if err := validateIsNotNow(exprs[0]); err != nil {
		panic(err)
	}

	err := validateModifiers(exprs[1:])
	if err != nil {
		panic(err)
	}

	return d.buildWithInputs(exprs)
}

// buildWithInputs builds the datetime function with the given inputs
func (d *DateTimeFunction) buildWithInputs(exprs []Expression) string {
	return d.buildFunctionString(func(stmt *sqlwriter.SqlWriter) {
		stmt.WriteList(len(exprs), func(i int) {
			stmt.WriteString(exprs[i].ToSQL())
		})
	})
}

// stringStrfTime is a special case, since it takes a format string as the first argument
func (d *DateTimeFunction) stringStrftime(exprs []Expression) string {
	// first argument must be a string
	if _, ok := exprs[0].(*ExpressionLiteral); !ok {
		panic("first argument to strftime must be a string")
	}

	format := exprs[0].ToSQL()

	err := parseFormat(format)
	if err != nil {
		panic(err)
	}

	if err := validateIsNotNow(exprs[1]); err != nil {
		panic(err)
	}

	// now we need validate the modifiers
	// the second input is the time value, so we need to validate the modifiers starting at the third input
	err = validateModifiers(exprs[2:])
	if err != nil {
		panic(err)
	}

	return d.buildWithInputs(exprs)
}

// validateIsNotNow validates that the input is not the ExpressionLiteral 'now'
func validateIsNotNow(exp Expression) error {
	literal, ok := exp.(*ExpressionLiteral)
	if !ok {
		return nil
	}

	value := trimLiteralQuotes(literal.Value)

	if strings.EqualFold(value, "now") {
		return fmt.Errorf("cannot use 'now' as an input for datetime function")
	}
	return nil
}

// validateModifiers validates that the modifiers are valid for the datetime function
func validateModifiers(exprs []Expression) error {
	for _, expr := range exprs {
		err := validateModifier(expr)
		if err != nil {
			return err
		}
	}
	return nil
}

// validateModifier validates that the modifier is valid for the datetime function
func validateModifier(expr Expression) error {
	literal, ok := expr.(*ExpressionLiteral)
	if !ok {
		return fmt.Errorf("datetime modifier must be an ExpressionLiteral")
	}

	// now we need to parse the modifier
	modifier := strings.Trim(literal.Value, " ")
	splitLen := len(strings.Split(modifier, " "))
	if !isStringLiteral(modifier) {
		return fmt.Errorf("modifier must be a string literal.  found: %s", modifier)
	}
	modifier = trimLiteralQuotes(modifier)

	if splitLen == 1 {
		if strings.EqualFold(modifier, "unixepoch") {
			return nil
		} else {
			return fmt.Errorf("modifier %s is not supported", modifier)
		}
	} else if splitLen == 2 {
		// can either be weekday N or NNN month/year/day
		if isValidWeekdayModifier(modifier) {
			return nil
		} else if isValidNNNModifier(modifier) {
			return nil
		} else {
			return fmt.Errorf("modifier %s is not supported", modifier)
		}
	} else if splitLen == 3 {
		// must be NNN.NNNN seconds
		if isValidStartOfModifier(modifier) {
			return nil
		} else {
			return fmt.Errorf("modifier %s is not supported", modifier)
		}
	}
	return fmt.Errorf("modifier %s is not supported", modifier)
}

func trimLiteralQuotes(literal string) string {
	literal = strings.Trim(literal, "'")
	return literal
}

func isValidStartOfModifier(modifier string) bool {
	modifier = strings.Trim(modifier, " ")
	if strings.EqualFold(modifier, "start of month") {
		return true
	} else if strings.EqualFold(modifier, "start of year") {
		return true
	} else if strings.EqualFold(modifier, "start of day") {
		return true
	} else {
		return false
	}
}

func isValidWeekdayModifier(modifier string) bool {
	modifier = strings.Trim(modifier, " ")
	splitModifier := strings.Split(modifier, " ")
	if len(splitModifier) != 2 {
		return false
	}

	if !strings.EqualFold(splitModifier[0], "weekday") {
		return false
	}

	// second part must be a number
	if _, err := strconv.Atoi(splitModifier[1]); err != nil {
		return false
	}

	return true
}

// isValidNNNModifier returns true if the string is a valid NNN modifier
// for example, 1 day, 2 days, 1 month, 2 months, 1 year, 2 years, 1 hour, 2 hours, 1 minute, 2 minutes, 1 second, 2 seconds are all valid
func isValidNNNModifier(modifier string) bool {
	modifier = strings.Trim(modifier, " ")
	splitModifier := strings.Split(modifier, " ")
	if len(splitModifier) != 2 {
		return false
	}

	// first part must be a number
	if !isValidNumberIncrement(splitModifier[0]) {
		return false
	}

	// second part must be a valid time denomination
	if !isValidTimeDenomination(splitModifier[1]) {
		return false
	}

	return true
}

// isValidTimeDenomination returns true if the string is a valid time denomination
// for example, "day", "days", "month", "months", "year", "years", "hour", "hours", "minute", "minutes", "second", "seconds" are all valid
func isValidTimeDenomination(denomination string) bool {
	denomination = strings.Trim(denomination, " ")
	switch denomination {
	case "day", "days", "month", "months", "year", "years", "hour", "hours", "minute", "minutes", "second", "seconds":
		return true
	default:
		return false
	}
}

// isValidNumberIncrement returns true if the string is a valid number increment
// for example, +1, -1, 1, 2, +2, -2, etc are all valid
func isValidNumberIncrement(number string) bool {
	number = strings.Trim(number, " ")
	if strings.HasPrefix(number, "+") || strings.HasPrefix(number, "-") {
		number = number[1:]
	}

	_, err := strconv.Atoi(number)
	return err == nil
}

// containsAny returns true if the string contains any of the substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

var unsupportedSubstitutions = []string{"%f", "%J", "%s"}

// parseFormat parses the format string for strftime to ensure that it is valid and does not contain any unsupported
// substitutions
func parseFormat(format string) error {
	if containsAny(format, unsupportedSubstitutions) {
		return fmt.Errorf("format string %s contains unsupported substitutions", format)
	}

	// check if it uses decimals
	if strings.Contains(format, ".") {
		return fmt.Errorf("format string %s contains unsupported decimals", format)
	}

	return nil
}
