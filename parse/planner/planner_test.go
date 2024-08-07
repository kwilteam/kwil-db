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
		vars    map[string]*types.DataType            // variables that can be accessed, can be nil
		objects map[string]map[string]*types.DataType // objects that can be referenced, can be nil
		err     error                                 // can be nil if no error is expected
	}

	tests := []testcase{
		{
			name: "basic select",
			sql:  "select 1",
			wt: "Projection: 1\n" +
				"└-Empty Scan\n",
		},
		{
			name: "array and object",
			sql:  "select $a.b, $c[1]",
			vars: map[string]*types.DataType{
				"$c": types.ArrayType(types.IntType),
			},
			objects: map[string]map[string]*types.DataType{
				"$a": {"b": types.IntType},
			},
			wt: "Projection: $a.b, $c[1]\n" +
				"└-Empty Scan\n",
		},
		{
			name: "select with filter",
			sql:  "select id, name from users where age > 18",
			wt: "Projection: users.id, users.name\n" +
				"└-Filter: users.age > 18\n" +
				"  └-Scan Table [alias=users]: users\n",
		},
		{
			name: "subquery join",
			sql:  "select name from users u inner join (select owner_id from posts) p on u.id = p.owner_id",
			wt: "Projection: u.name\n" +
				"└-Join [inner]: u.id = p.owner_id\n" +
				"  |-Scan Table [alias=u]: users\n" +
				"  └-Scan Subquery [alias=p]:\n" +
				"    └-Projection: posts.owner_id\n" +
				"      └-Scan Table [alias=posts]: posts\n",
		},
		{
			name: "scalar subquery in where clause",
			sql:  "select name from users where id = (select id from posts where content = 'hello')",
			wt: "Projection: users.name\n" +
				"└-Filter: users.id = subquery [regular] [subplan_id=0]\n" +
				"  └-Scan Table [alias=users]: users\n" +
				"Subplan [id=0]: [scalar]\n" +
				"└-Projection: posts.id\n" +
				"  └-Filter: posts.content = 'hello'\n" +
				"    └-Scan Table [alias=posts]: posts\n",
		},
		{
			name: "correlated subquery in where clause",
			sql:  "select name from users u where exists (select 1 from posts p where p.owner_id = u.id)",
			wt: "Projection: u.name\n" +
				"└-Filter: subquery [exists] [subplan_id=0]\n" +
				"  └-Scan Table [alias=u]: users\n" +
				"Subplan [id=0]: [correlated=u.id]\n" +
				"└-Projection: 1\n" +
				"  └-Filter: p.owner_id = u.id\n" +
				"    └-Scan Table [alias=p]: posts\n",
		},
		{
			// tests that correlation is propagated across multiple subqueries
			name: "double nested correlated subquery",
			sql:  "select name from users u where exists (select 1 from posts p where exists (select 1 from posts p2 where p2.owner_id = u.id))",
			wt: "Projection: u.name\n" +
				"└-Filter: subquery [exists] [subplan_id=1]\n" +
				"  └-Scan Table [alias=u]: users\n" +
				"Subplan [id=1]: [correlated=u.id]\n" +
				"└-Projection: 1\n" +
				"  └-Filter: subquery [exists] [subplan_id=0]\n" +
				"    └-Scan Table [alias=p]: posts\n" +
				"Subplan [id=0]: [correlated=u.id]\n" +
				"└-Projection: 1\n" +
				"  └-Filter: p2.owner_id = u.id\n" +
				"    └-Scan Table [alias=p2]: posts\n",
		},
		{
			name: "aggregate without group by",
			sql:  "select sum(age) from users",
			wt: "Projection: sum(users.age)\n" +
				"└-Aggregate: sum(users.age)\n" +
				"  └-Scan Table [alias=users]: users\n",
		},
		{
			name: "aggregate with group by",
			sql:  "select name, sum(age) from users group by name having sum(age)::int > 100",
			wt: "Projection: users.name, sum(users.age)\n" +
				"└-Filter: sum(users.age)::int > 100\n" +
				"  └-Aggregate [users.name]: sum(users.age)\n" +
				"    └-Scan Table [alias=users]: users\n",
		},
		{
			name: "complex group by",
			sql:  "select age/2, age*3 from users group by age/2, age*3",
			wt: "Projection: users.age / 2, users.age * 3\n" +
				"└-Aggregate [users.age / 2] [users.age * 3]: \n" +
				"  └-Scan Table [alias=users]: users\n",
		},
		// TODO: negative case of the above
		{
			name: "complex having",
			sql:  "select name, sum(age/2)+sum(age*10) from users group by name having sum(age)::int > 100 or sum(age/2)::int > 10",
			wt: "Projection: users.name, sum(users.age / 2) + sum(users.age * 10)\n" +
				"└-Filter: (sum(users.age)::int > 100 OR sum(users.age / 2)::int > 10)\n" +
				"  └-Aggregate [users.name]: sum(users.age / 2), sum(users.age * 10), sum(users.age)\n" +
				"    └-Scan Table [alias=users]: users\n",
		},
		// TODO: test that we cannot use aggregates in where clause
		{
			name: "every type of join",
			vars: map[string]*types.DataType{
				"$id":   types.IntType,
				"$name": types.TextType,
			},
			sql: `select c.brand, pu.content, u.name, u2.id, count(p.id) from users u
				inner join posts p on u.id = p.owner_id
				left join owned_cars['dbid', 'proc']($id) c on c.owner_name = u.name
				right join posts_by_user($name) pu on pu.content = p.content
				full join (select id from users where age > 18) u2 on u2.id = u.id
				group by c.brand, pu.content, u.name, u2.id;`,
			wt: "Projection: c.brand, pu.content, u.name, u2.id, count(p.id)\n" +
				"└-Aggregate [c.brand] [pu.content] [u.name] [u2.id]: count(p.id)\n" +
				"  └-Join [outer]: u2.id = u.id\n" +
				"    |-Join [right]: pu.content = p.content\n" +
				"    | |-Join [left]: c.owner_name = u.name\n" +
				"    | | |-Join [inner]: u.id = p.owner_id\n" +
				"    | | | |-Scan Table [alias=u]: users\n" +
				"    | | | └-Scan Table [alias=p]: posts\n" +
				"    | | └-Scan Procedure [alias=c]: [foreign=true] [dbid='dbid'] [proc='proc'] owned_cars($id)\n" +
				"    | └-Scan Procedure [alias=pu]: [foreign=false] posts_by_user($name)\n" +
				"    └-Scan Subquery [alias=u2]:\n" +
				"      └-Projection: users.id\n" +
				"        └-Filter: users.age > 18\n" +
				"          └-Scan Table [alias=users]: users\n",
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

				// special case for testing
				if errors.Is(test.err, errAny) {
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

// special error for testing that will match any error
var errAny = errors.New("any error")

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

procedure posts_by_user($name text) public view returns table(content text) {
	return select p.content from posts p
		inner join users u on p.owner_id = u.id
		where u.name = $name;
}

foreign procedure owned_cars($id int) returns table(owner_name text, brand text, model text)
`
