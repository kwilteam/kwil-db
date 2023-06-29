package actparser

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"testing"
)

var trace = flag.Bool("trace", false, "run tests with tracing")

func TestParseActionStmt(t *testing.T) {

	tests := []struct {
		name   string
		input  string
		expect ActionStmt
	}{
		{
			name:  "action_call",
			input: `action_xx($a, 2, "3");`,
			expect: &ActionCallStmt{
				Method: "action_xx",
				Args: []string{
					`$a`,
					`2`,
					`"3"`,
				},
			},
		},
		{
			name:  "extension_call",
			input: `$a, $b = erc20.transfer($q, 2, "3");`,
			expect: &ExtensionCallStmt{
				Extension: "erc20",
				Method:    "transfer",
				Args: []string{
					`$q`,
					`2`,
					`"3"`,
				},
				Receivers: []string{
					`$a`,
					`$b`,
				},
			},
		},
		{
			name:  "dml select",
			input: `SELECT * FROM users;`,
			expect: &DMLStmt{
				Statement: `SELECT * FROM users;`,
			},
		},
		{
			name:  "dml insert",
			input: `insert into users (id, name) values (1, "test");`,
			expect: &DMLStmt{
				Statement: `insert into users (id, name) values (1, "test");`,
			},
		},
		{
			name:  "dml update",
			input: `update users set name = "test" where id = 1;`,
			expect: &DMLStmt{
				Statement: `update users set name = "test" where id = 1;`,
			},
		},
		{
			name:  "dml delete",
			input: `delete from users where id = 1;`,
			expect: &DMLStmt{
				Statement: `delete from users where id = 1;`,
			},
		},
		{
			name:  "dml with",
			input: `with x as (select * from users) select * from x;`,
			expect: &DMLStmt{
				Statement: `with x as (select * from users) select * from x;`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAst, err := ParseActionStmt(tt.input, nil, *trace)
			if err != nil {
				t.Errorf("ParseActionStmt() error = %v", err)
				return
			}

			assert.EqualValues(t, tt.expect, gotAst, "ParseRawSQL() got %+v, want %+v", gotAst, tt.expect)
		})
	}
}
