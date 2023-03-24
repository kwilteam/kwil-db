package sql

import (
	"flag"
	"kwil/internal/pkg/kl/types"
	"strings"
	"testing"
)

var trace = flag.Bool("trace", false, "run tests with tracing")

func TestParseRawSQL_banRules(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError error
	}{
		// non-deterministic time functions
		{"select date function", "select date('2020-02-02')", nil},
		{"select date function 1", "select date('now')", ErrFunctionNotSupported},
		{"select date function 2", "select date('now', '+1 day')", ErrFunctionNotSupported},
		{"select time function", "select time('now')", ErrFunctionNotSupported},
		{"select datetime function", "select datetime('now')", ErrFunctionNotSupported},
		{"select julianday function", "select julianday('now')", ErrFunctionNotSupported},
		{"select unixepoch function", "select unixepoch('now')", ErrFunctionNotSupported},
		{"select strftime function", "select strftime('%Y%m%d', 'now')", ErrFunctionNotSupported},
		{"select strftime function 2", "select strftime('%Y%m%d')", ErrFunctionNotSupported},
		//
		{"upsert date function", "insert into t1 values (date('now'))", ErrFunctionNotSupported},
		{"upsert date function 2", "insert into t1 values ( date('now', '+1 day'))", ErrFunctionNotSupported},
		{"upsert time function", "insert into t1 values ( time('now'))", ErrFunctionNotSupported},
		{"upsert datetime function", "insert into t1 values ( datetime('now'))", ErrFunctionNotSupported},
		{"upsert julianday function", "insert into t1 values ( julianday('now'))", ErrFunctionNotSupported},
		{"upsert unixepoch function", "insert into t1 values ( unixepoch('now'))", ErrFunctionNotSupported},
		{"upsert strftime function", "insert into t1 values (strftime('%Y%m%d', 'now'))", ErrFunctionNotSupported},
		{"upsert strftime function 2", "insert into t1 values (strftime('%Y%m%d'))", ErrFunctionNotSupported},
		// non-deterministic random functions
		{"random function", "select random()", ErrFunctionNotSupported},
		{"randomblob function", "select randomblob(10)", ErrFunctionNotSupported},
		{"random function", "insert into t2 values ( random())", ErrFunctionNotSupported},
		{"randomblob function", "insert into t2 values ( randomblob(10))", ErrFunctionNotSupported},
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
		// non-exist table/column
		{"select from table", "select * from t1", nil},
		{"select with CTE", "with tt as (select * from t1) select * from tt", nil},
		{"select non-exist table", "select * from t10", ErrTableNotFound},
		// joins
		{"joins", "select * from t1 join t2 on t1.c1=t2.c1 ", nil},
		{"joins 3", "select * from t1 join t2 on (1+2)=2", ErrJoinConditionTooDeep},
		{"joins 2", "select * from t1 join t2 on t1.c1=t7.c1 ", ErrTableNotFound},
		{"joins 3", "select * from t1 join t2 on t1.c1=t2.c7", ErrColumnNotFound},

		{"multi joins", "select * from t1 join t2 join t3 join t4 on a=b", nil},
		{"multi joins 2", "select * from t1 join t2 join t3 join t4 join t5 on a=b", ErrMultiJoinNotSupported},
		{"natural join", "select * from t3 natural join t4", ErrJoinNotSupported},
		{"cross join", "select * from t3 cross join t4", ErrJoinNotSupported},
		{"cartesian join 1", "select * from t3, t4", ErrSelectFromMultipleTables},
		// join without any condition
		{"cartesian join 2 1", "select * from t3 left join t4", ErrJoinWithoutCondition},
		{"cartesian join 2 2", "select * from t3 inner join t4", ErrJoinWithoutCondition},
		{"cartesian join 2 3", "select * from t3 join t4", ErrJoinWithoutCondition},
		// join with condition
		{"join with unary cons", "select * from t3 join t4 on not a", ErrJoinConditionOpNotSupported},
		{"join with non = cons", "select * from t3 join t4 on a and b", ErrJoinConditionOpNotSupported},
		{"join with non = cons 2", "select * from t3 join t4 on a + b", ErrJoinConditionOpNotSupported},
		//{"join with multi level binary cons", "select * from t3 join t4 on a=(b=c)", nil}, // TODO
		//{"join with function cons", "select * from t3 join t4 on random()", ErrJoinWithTrueCondition}, /// TODO: support this
		// action parameters
		{"insert with bind parameter", "insert into t3 values ($this)", nil},
		{"insert with non exist bond parameter", "insert into t3 values ($a)", ErrBindParameterNotFound},
		// modifiers
		{"modifier", "select * from t3 where a = @caller", nil},
		{"modifier 2", "select * from t3 where a = @block_height", nil},
		{"modifier 3", "select * from t3 where a = @any", ErrModifierNotSupported},
	}

	ctx := types.DatabaseContext{
		Tables: map[string]types.TableContext{
			"t1": {
				Columns:     []string{"c1", "c2", "c3"},
				PrimaryKeys: []string{"c1"},
			},
			"t2": {
				Columns:     []string{"c1", "c2", "c3"},
				PrimaryKeys: []string{"c1"},
			},
			"t3": {
				Columns:     []string{"c1", "c2", "c3"},
				PrimaryKeys: []string{"c1"},
			},
			"t4": {
				Columns:     []string{"c1", "c2", "c3"},
				PrimaryKeys: []string{"c1"},
			},
			"t5": {},
		},
		Actions: map[string]types.ActionContext{
			"action1": {
				"this": nil,
				"that": nil,
			},
			"action2": {
				"here":  nil,
				"there": nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ParseRawSQL(test.input, 1, "action1", ctx, *trace)

			if err == nil && test.wantError == nil {
				return
			}

			if err != nil && test.wantError != nil {
				// TODO: errors.Is?
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
