package planner_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/kwilteam/kwil-db/parse/planner"
	"github.com/stretchr/testify/require"
)

func Test_Planner(t *testing.T) {
	type testcase struct {
		name    string
		sql     string
		wt      string                                // want, abbreviated for formatting test cases
		vars    map[string]*types.DataType            // can be nil if no vars are expected
		objects map[string]map[string]*types.DataType // can be nil if no objects are expected
		err     error                                 // can be nil if no error is expected
	}

	tests := []testcase{
		{
			name: "basic select",
			sql:  "select 1",
			wt: "Projection: 1\n" +
				"  Empty Scan\n",
		},
		{
			name: "select with filter",
			sql:  "select id, name from users where age > 18",
			wt: "Projection: users.id, users.name\n" +
				"  Filter: users.age > 18\n" +
				"    Scan Table [alias=users]: users\n",
		},
		{
			name: "subquery join",
			sql:  "select name from users u inner join (select owner_id from posts) p on u.id = p.owner_id",
			wt: "Projection: u.name\n" +
				"  Inner Join: u.id = p.owner_id\n" +
				"    Scan Table [alias=u]: users\n" +
				"    Scan Subquery [alias=p]: \n" +
				"      Projection: posts.owner_id\n" +
				"        Scan Table [alias=posts]: posts\n",
		},
		{
			name: "scalar subquery in where clause",
			sql:  "select name from users where id = (select id from posts where content = 'hello')",
			wt: "Projection: users.name\n" +
				"  Filter: users.id = subquery [regular] [subplan_id=0]\n" +
				"    Scan Table [alias=users]: users\n" +
				"Subplan [id=0]\n" +
				"  Projection: posts.id\n" +
				"    Filter: posts.content = 'hello'\n" +
				"      Scan Table [alias=posts]: posts\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := parse.Parse([]byte(testSchema))
			require.NoError(t, err)

			parsedSql, err := parse.ParseSQL(test.sql, schema, true)
			require.NoError(t, err)
			require.NoError(t, parsedSql.ParseErrs.Err())

			plan, err := planner.Plan(parsedSql.AST, schema, test.vars, test.objects)
			if test.err != nil {
				require.Error(t, err)

				if errors.Is(err, anyErr) {
					return
				}

				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)

				// TODO: delete this block once I am done debugging
				rec := planner.Format(plan)
				if test.wt != rec {
					fmt.Println(rec)
					require.Equal(t, test.wt, planner.Format(plan))
				}
				// TODO: end delete here

				require.Equal(t, test.wt, planner.Format(plan))
			}
		})
	}
}

// special error for testing
var anyErr = errors.New("any error")

var testSchema = `database planner;

table users {
	id uuid primary key,
	name text,
	age int max(150)
}

table posts {
	id uuid primary key,
	owner_id uuid not null,
	content text maxlen(300),
	foreign key (owner_id) references users(id) on delete cascade on update cascade
}
`
