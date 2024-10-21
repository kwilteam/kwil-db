package optimizer

import (
	"errors"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/kwilteam/kwil-db/parse/planner/logical"
	"github.com/stretchr/testify/require"
)

func Test_Pushdown(t *testing.T) {
	type testcase struct {
		name string
		sql  string
		wt   string // expected logical plan, uses name "wt" for formatting
		err  error
	}

	tests := []testcase{
		{
			name: "simple pushdown",
			sql:  "select * from users where name = 'foo'",
			wt: "Return: id [uuid], name [text], age [int]\n" +
				"└─Project: users.id; users.name; users.age\n" +
				"  └─Scan Table: users [physical] filter=[users.name = 'foo']\n",
		},
		{
			name: "push down join",
			sql:  "select * from users u inner join posts p on u.id = p.owner_id where u.name = p.content",
			wt: "Return: id [uuid], name [text], age [int], id [uuid], owner_id [uuid], content [text], created_at [int]\n" +
				"└─Project: u.id; u.name; u.age; p.id; p.owner_id; p.content; p.created_at\n" +
				"  └─Join [inner]: u.id = p.owner_id AND u.name = p.content\n" +
				"    ├─Scan Table [alias=\"u\"]: users [physical]\n" +
				"    └─Scan Table [alias=\"p\"]: posts [physical]\n",
		},
		{
			name: "pushdown through join",
			sql:  "select * from users u inner join posts p on u.id = p.owner_id where u.name = 'foo'",
			wt: "Return: id [uuid], name [text], age [int], id [uuid], owner_id [uuid], content [text], created_at [int]\n" +
				"└─Project: u.id; u.name; u.age; p.id; p.owner_id; p.content; p.created_at\n" +
				"  └─Join [inner]: u.id = p.owner_id\n" +
				"    ├─Scan Table [alias=\"u\"]: users [physical] filter=[u.name = 'foo']\n" +
				"    └─Scan Table [alias=\"p\"]: posts [physical]\n",
		},
		{
			name: "doesn't pushdown through join if not all columns are from one side",
			sql:  "select u.id from users u inner join posts p on u.id = p.owner_id inner join users u2 on u2.id = p.owner_id where u.age + u2.age = length(p.content)",
			wt: "Return: id [uuid]\n" +
				"└─Project: u.id\n" +
				"  └─Join [inner]: u2.id = p.owner_id AND u.age + u2.age = length(p.content)\n" +
				"    ├─Join [inner]: u.id = p.owner_id\n" +
				"    │ ├─Scan Table [alias=\"u\"]: users [physical]\n" +
				"    │ └─Scan Table [alias=\"p\"]: posts [physical]\n" +
				"    └─Scan Table [alias=\"u2\"]: users [physical]\n",
		},
		{
			name: "split and pushdown AND",
			sql:  "select u.id from users u inner join posts p on u.id = p.owner_id where u.age = 10 and p.created_at = 100",
			wt: "Return: id [uuid]\n" +
				"└─Project: u.id\n" +
				"  └─Join [inner]: u.id = p.owner_id\n" +
				"    ├─Scan Table [alias=\"u\"]: users [physical] filter=[u.age = 10]\n" +
				"    └─Scan Table [alias=\"p\"]: posts [physical] filter=[p.created_at = 100]\n",
		},
		{
			// shouldn't do anything, just checking that aggregates work as expected.
			// Since the planner uses a reference system for any aggregate results,
			// there is no need to push down any aggregate function.
			name: "aggregate having",
			sql:  "select u.id, count(p.id) from users u inner join posts p on u.id = p.owner_id group by u.id having count(p.id) > 10",
			wt: "Return: id [uuid], count [int]\n" +
				"└─Project: {#ref(A)}; {#ref(B)}\n" +
				"  └─Filter: {#ref(B)} > 10\n" +
				"    └─Aggregate [{#ref(A) = u.id}]: {#ref(B) = count(p.id)}\n" +
				"      └─Join [inner]: u.id = p.owner_id\n" +
				"        ├─Scan Table [alias=\"u\"]: users [physical]\n" +
				"        └─Scan Table [alias=\"p\"]: posts [physical]\n",
		},
		{
			name: "update with a basic FROM clause",
			sql:  "update users set name = 'foo' from posts where users.id = posts.owner_id",
			wt: "Update [users]: name = 'foo'\n" +
				"└─Join [inner]: users.id = posts.owner_id\n" +
				"  ├─Scan Table: users [physical]\n" +
				"  └─Scan Table: posts [physical]\n",
		},
		{
			name: "more complex update with a FROM clause",
			sql:  "update users set name = posts.content from posts inner join users u on u.id = posts.owner_id where users.id = u.id",
			wt: "Update [users]: name = posts.content\n" +
				"└─Join [inner]: users.id = u.id\n" +
				"  ├─Scan Table: users [physical]\n" +
				"  └─Join [inner]: u.id = posts.owner_id\n" +
				"    ├─Scan Table: posts [physical]\n" +
				"    └─Scan Table [alias=\"u\"]: users [physical]\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := parse.Parse([]byte(testSchema))
			require.NoError(t, err)

			parsedSql, err := parse.ParseSQL(test.sql, schema, true)
			require.NoError(t, err)
			require.NoError(t, parsedSql.ParseErrs.Err())

			plan, err := logical.CreateLogicalPlan(parsedSql.AST, schema, map[string]*types.DataType{}, map[string]map[string]*types.DataType{}, false)
			require.NoError(t, err)

			newPlan, err := PushdownPredicates(plan.Plan)
			plan.Plan = newPlan
			if test.err != nil {
				require.Error(t, err)

				// special case for testing
				if errors.Is(test.err, errAny) {
					return
				}

				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.wt, plan.Format())
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
