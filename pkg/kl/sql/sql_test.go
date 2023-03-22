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
		// non-deterministic functions in select
		{"date function", "select date('2020-02-02')", nil},
		{"date function", "select date('now')", nil},
		{"date function", "select date('now', '+1 day')", nil},
		{"time function", "select time('now')", nil},
		{"datetime function", "select datetime('now')", nil},
		{"julianday function", "select julianday('now')", nil},
		{"unixepoch function", "select unixepoch('now')", nil},
		{"random function", "select random()", nil},
		{"randomblob function", "select randomblob(10)", nil},
		// non-deterministic functions in insert update
		{"date function", "insert into t values (date('now'))", ErrFunctionNotSupported},
		{"date function", "insert into t values ( date('now', '+1 day'))", ErrFunctionNotSupported},
		{"time function", "insert into t values ( time('now'))", ErrFunctionNotSupported},
		{"datetime function", "insert into t values ( datetime('now'))", ErrFunctionNotSupported},
		{"julianday function", "insert into t values ( julianday('now'))", ErrFunctionNotSupported},
		{"unixepoch function", "insert into t values ( unixepoch('now'))", ErrFunctionNotSupported},
		{"random function", "insert into t values ( random())", ErrFunctionNotSupported},
		{"randomblob function", "insert into t values ( randomblob(10))", ErrFunctionNotSupported},
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
			// nil error
			if err == test.wantError {
				return
			}

			if strings.Contains(err.Error(), test.wantError.Error()) {
				return
			}
			t.Errorf("ParseRawSQL() expected error: %s, got %s", test.wantError, err)
		})
	}
}
