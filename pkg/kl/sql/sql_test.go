package sql

import (
	"kwil/pkg/kl/ast"
	"strings"
	"testing"
)

func TestParseRawSQL_banRules(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError error
	}{
		// non-deterministic time functions
		{"select date function", "select date('2020-02-02')", nil},
		{"select date function", "select date('now')", ErrFunctionNotSupported},
		{"select date function", "select date('now', '+1 day')", ErrFunctionNotSupported},
		{"select time function", "select time('now')", ErrFunctionNotSupported},
		{"select datetime function", "select datetime('now')", ErrFunctionNotSupported},
		{"select julianday function", "select julianday('now')", ErrFunctionNotSupported},
		{"select unixepoch function", "select unixepoch('now')", ErrFunctionNotSupported},
		{"select strftime function", "select strftime('%Y%m%d', 'now')", ErrFunctionNotSupported},
		{"select strftime function 2", "select strftime('%Y%m%d')", ErrFunctionNotSupported},
		//
		{"upsert date function", "insert into t values (date('now'))", ErrFunctionNotSupported},
		{"upsert date function", "insert into t values ( date('now', '+1 day'))", ErrFunctionNotSupported},
		{"upsert time function", "insert into t values ( time('now'))", ErrFunctionNotSupported},
		{"upsert datetime function", "insert into t values ( datetime('now'))", ErrFunctionNotSupported},
		{"upsert julianday function", "insert into t values ( julianday('now'))", ErrFunctionNotSupported},
		{"upsert unixepoch function", "insert into t values ( unixepoch('now'))", ErrFunctionNotSupported},
		{"upsert strftime function", "insert into t values (strftime('%Y%m%d', 'now'))", ErrFunctionNotSupported},
		{"upsert strftime function 2", "insert into t values (strftime('%Y%m%d'))", ErrFunctionNotSupported},
		// non-deterministic random functions
		{"random function", "select random()", ErrFunctionNotSupported},
		{"randomblob function", "select randomblob(10)", ErrFunctionNotSupported},
		{"random function", "insert into t values ( random())", ErrFunctionNotSupported},
		{"randomblob function", "insert into t values ( randomblob(10))", ErrFunctionNotSupported},
		// non-deterministic math functions
		{"select acos function", "select acos(1)", ErrFunctionNotSupported},
		{"select acosh function", "select acosh(1)", ErrFunctionNotSupported},
		{"select asin function", "select asin(1)", ErrFunctionNotSupported},
		{"select asinh function", "select asinh(1)", ErrFunctionNotSupported},
		{"select atan function", "select atan(1)", ErrFunctionNotSupported},
		{"select atan2 function", "select atan2(1, 1)", ErrFunctionNotSupported},
		{"select atanh function", "select atanh(1)", ErrFunctionNotSupported},
		{"select ceil function", "select ceil(1)", ErrFunctionNotSupported},
		{"select ceiling function", "select ceiling(1)", ErrFunctionNotSupported},
		{"select cos function", "select cos(1)", ErrFunctionNotSupported},
		{"select cosh function", "select cosh(1)", ErrFunctionNotSupported},
		{"select degrees function", "select degrees(1)", ErrFunctionNotSupported},
		{"select exp function", "select exp(1)", ErrFunctionNotSupported},
		{"select ln function", "select ln(1)", ErrFunctionNotSupported},
		{"select log function", "select log(1)", ErrFunctionNotSupported},
		{"select log function 2", "select log(1, 1)", ErrFunctionNotSupported},
		{"select log10 function", "select log10(1)", ErrFunctionNotSupported},
		{"select log2 function", "select log2(1)", ErrFunctionNotSupported},
		{"select mod function", "select mod(1, 1)", ErrFunctionNotSupported},
		{"select pi function", "select pi()", ErrFunctionNotSupported},
		{"select pow function", "select pow(1, 1)", ErrFunctionNotSupported},
		{"select power function", "select power(1, 1)", ErrFunctionNotSupported},
		{"select radians function", "select radians(1)", ErrFunctionNotSupported},
		{"select sin function", "select sin(1)", ErrFunctionNotSupported},
		{"select sinh function", "select sinh(1)", ErrFunctionNotSupported},
		{"select sqrt function", "select sqrt(1)", ErrFunctionNotSupported},
		{"select tan function", "select tan(1)", ErrFunctionNotSupported},
		{"select tanh function", "select tanh(1)", ErrFunctionNotSupported},
		{"select trunc function", "select trunc(1)", ErrFunctionNotSupported},
		// non-deterministic keywords
		{"current_date", "select current_date", ErrKeywordNotSupported},
		{"current_time", "select current_time", ErrKeywordNotSupported},
		{"current_timestamp", "select current_timestamp", ErrKeywordNotSupported},
		// joins
		{"multi joins", "select * from users join posts join comments on a=b", ErrMultiJoinNotSupported},
		{"natural join", "select * from users natural join posts", ErrJoinNotSupported},
		{"cross join", "select * from users cross join posts", ErrJoinNotSupported},
		{"cartesian join 1", "select * from users, posts", ErrSelectFromMultipleTables},
		// join without any condition
		{"cartesian join 2 1", "select * from users left join posts", ErrJoinWithoutCondition},
		{"cartesian join 2 2", "select * from users inner join posts", ErrJoinWithoutCondition},
		{"cartesian join 2 3", "select * from users join posts", ErrJoinWithoutCondition},
		// join with condition
		//{"join with multi level binary cons", "select * from users join posts on a=(b+c)", nil},
		{"join with unary cons", "select * from users join posts on not a", ErrJoinConditionOpNotSupported},
		{"join with non = cons", "select * from users join posts on a and b", ErrJoinConditionOpNotSupported},
		//{"join with function cons", "select * from users join posts on random()", ErrJoinWithTrueCondition}, /// TODO: support this
		// action parameters
		{"insert", "insert into t values ($this)", nil},
		{"insert", "insert into t values ($a)", ErrBindParameterNotFound},
	}

	ctx := ast.ActionContext{
		"this": nil,
		"that": nil,
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ParseRawSQL(test.input, 1, ctx, false)
			if err == nil && test.wantError == nil {
				return
			}

			if err != nil && test.wantError != nil {
				if strings.Contains(err.Error(), test.wantError.Error()) {
					return
				}
				t.Errorf("ParseRawSQL() expected error: %s, got %s", test.wantError, err)
				return
			}

			t.Errorf("ParseRawSQL() expected: %s, got %s", test.wantError, err)
		})
	}
}
