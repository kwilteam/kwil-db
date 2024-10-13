package interpreter_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/parse"
	"github.com/kwilteam/kwil-db/parse/interpreter"
	"github.com/stretchr/testify/require"
)

func Test_Interpeter(t *testing.T) {
	type testcase struct {
		name      string
		procName  string  // name of the procedure, must match the definition below in the proc field
		proc      string  // procedure definition in Kuneiform
		inputVals []any   // input values for the procedure, can be nil
		expected  [][]any // can be nil if error is expected or if the procedure returns nothing
		err       error   // can be nil
	}

	tests := []testcase{
		{
			name:     "assign a variable",
			procName: "assign",
			proc: `
			procedure assign($a int) public {
				$b := $a;
			}
			`,
			inputVals: []any{int64(5)},
		},
		{
			name:     "return a variable",
			procName: "returnme",
			proc: `
			procedure returnme($a int) public returns (id int) {
				return $a;
			}
			`,
			inputVals: []any{int64(5)},
			expected:  [][]any{{int64(5)}},
		},
		{
			name:     "return a table",
			procName: "posts_by_user2",
			proc: `
			procedure posts_by_user2() public view returns table(content text) {
				return next 'hello';
				return next 'world';
			}`,
			expected: [][]any{{"hello"}, {"world"}},
		},
		{
			name:     "loop",
			procName: "loop",
			proc: `
			procedure loop($start int, $end int) public returns (res int) {
				$res := 0;
				for $i in $start..$end {
					$res := $res + $i;
				}

				return $res;
			}
			`,
			inputVals: []any{int64(1), int64(5)},
			expected:  [][]any{{int64(15)}},
		},
		{
			name:     "loop and break",
			procName: "loop_and_break",
			proc: `
			procedure loop_and_break($start int, $end int) public returns (res int) {
				$res := 0;
				for $i in $start..$end {
					if $i == 3 {
						break;
					}
					$res := $res + $i;
				}
				
				return $res;
			}`,
			inputVals: []any{int64(1), int64(5)},
			expected:  [][]any{{int64(3)}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := parse.Parse([]byte(testSchema + test.proc))
			require.NoError(t, err)

			ctx := context.Background()

			proc, ok := schema.FindProcedure(test.procName)
			if !ok {
				t.Fatalf("procedure %s not found", test.procName)
			}

			res, err := interpreter.Run(ctx, proc, schema, test.inputVals)
			if test.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.err)
				return
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, len(test.expected), len(res.Values))
			for i, row := range res.Values {
				require.Equal(t, len(test.expected[i]), len(row))
				for j, val := range row {
					require.EqualValues(t, test.expected[i][j], val.Value.Value())
				}
			}
		})
	}
}

var testSchema = `database planner;

table users {
	id uuid primary key,
	name text,
	age int max(150),
	#name_idx unique(name),
	#age_idx index(age)
}

table posts {
	id uuid primary key,
	owner_id uuid not null,
	content text maxlen(300) unique,
	created_at int not null,
	foreign key (owner_id) references users(id) on delete cascade on update cascade,
	#owner_created_idx unique(owner_id, created_at)
}

procedure posts_by_user($name text) public view returns table(content text) {
	return select p.content from posts p
		inner join users u on p.owner_id = u.id
		where u.name = $name;
}

procedure post_count($id uuid) public view returns (int) {
	for $row in select count(*) as count from posts where owner_id = $id {
		return $row.count;
	}
}

foreign procedure owned_cars($id int) returns table(owner_name text, brand text, model text)
foreign procedure car_count($id uuid) returns (int)
`
