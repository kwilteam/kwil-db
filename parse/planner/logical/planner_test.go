package logical_test

import (
	"errors"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/kwilteam/kwil-db/parse/planner/logical"
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
			wt: "Return: ?column? [int]\n" +
				"└─Project: 1\n" +
				"  └─Empty Scan\n",
		},
		{
			name: "array and object",
			sql:  "select $a.b, $c as c1",
			vars: map[string]*types.DataType{
				"$c": types.ArrayType(types.IntType),
			},
			objects: map[string]map[string]*types.DataType{
				"$a": {"b": types.IntType},
			},
			wt: "Return: ?column? [int], c1 [int[]]\n" +
				"└─Project: $a.b; $c AS c1\n" +
				"  └─Empty Scan\n",
		},
		{
			name: "select with filter",
			sql:  "select id, name from users where age > 18",
			wt: "Return: id [uuid], name [text]\n" +
				"└─Project: users.id; users.name\n" +
				"  └─Filter: users.age > 18\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "subquery join",
			sql:  "select name from users u inner join (select owner_id from posts) p on u.id = p.owner_id",
			wt: "Return: name [text]\n" +
				"└─Project: u.name\n" +
				"  └─Join [inner]: u.id = p.owner_id\n" +
				"    ├─Scan Table [alias=\"u\"]: users [physical]\n" +
				"    └─Scan Subquery [alias=\"p\"]: [subplan_id=0] (uncorrelated)\n" +
				"Subplan [subquery] [id=0]\n" +
				"└─Project: posts.owner_id\n" +
				"  └─Scan Table: posts [physical]\n",
		},
		{
			name: "correlated joined subquery",
			sql:  "select name from users u where id = (select owner_id from posts inner join (select age from users where id = u.id) as u2 on u2.age=length(posts.content))",
			wt: "Return: name [text]\n" +
				"└─Project: u.name\n" +
				"  └─Filter: u.id = [subquery (scalar) (subplan_id=1) (correlated: u.id)]\n" +
				"    └─Scan Table [alias=\"u\"]: users [physical]\n" +
				"Subplan [subquery] [id=1]\n" +
				"└─Project: posts.owner_id\n" +
				"  └─Join [inner]: u2.age = length(posts.content)\n" +
				"    ├─Scan Table: posts [physical]\n" +
				"    └─Scan Subquery [alias=\"u2\"]: [subplan_id=0] (correlated: u.id)\n" +
				"Subplan [subquery] [id=0]\n" +
				"└─Project: users.age\n" +
				"  └─Filter: users.id = u.id\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "scalar subquery in where clause",
			sql:  "select name from users where id = (select id from posts where content = 'hello')",
			wt: "Return: name [text]\n" +
				"└─Project: users.name\n" +
				"  └─Filter: users.id = [subquery (scalar) (subplan_id=0) (uncorrelated)]\n" +
				"    └─Scan Table: users [physical]\n" +
				"Subplan [subquery] [id=0]\n" +
				"└─Project: posts.id\n" +
				"  └─Filter: posts.content = 'hello'\n" +
				"    └─Scan Table: posts [physical]\n",
		},
		{
			name: "correlated subquery in where clause",
			sql:  "select name from users u where exists (select 1 from posts p where p.owner_id = u.id)",
			wt: "Return: name [text]\n" +
				"└─Project: u.name\n" +
				"  └─Filter: [subquery (exists) (subplan_id=0) (correlated: u.id)]\n" +
				"    └─Scan Table [alias=\"u\"]: users [physical]\n" +
				"Subplan [subquery] [id=0]\n" +
				"└─Project: 1\n" +
				"  └─Filter: p.owner_id = u.id\n" +
				"    └─Scan Table [alias=\"p\"]: posts [physical]\n",
		},
		{
			name: "subquery in result",
			sql:  "select (select * from (select id from posts where owner_id = users.id) as p limit 1) from users",
			wt: "Return: id [uuid]\n" +
				"└─Project: [subquery (scalar) (subplan_id=1) (correlated: users.id)]\n" +
				"  └─Scan Table: users [physical]\n" +
				"Subplan [subquery] [id=1]\n" +
				"└─Project: p.id\n" +
				"  └─Limit: 1\n" +
				"    └─Scan Subquery [alias=\"p\"]: [subplan_id=0] (correlated: users.id)\n" +
				"Subplan [subquery] [id=0]\n" +
				"└─Project: posts.id\n" +
				"  └─Filter: posts.owner_id = users.id\n" +
				"    └─Scan Table: posts [physical]\n",
		},
		{
			name: "subquery exists",
			sql:  "select name from users where exists (select 1 from posts where owner_id = users.id)",
			wt: "Return: name [text]\n" +
				"└─Project: users.name\n" +
				"  └─Filter: [subquery (exists) (subplan_id=0) (correlated: users.id)]\n" +
				"    └─Scan Table: users [physical]\n" +
				"Subplan [subquery] [id=0]\n" +
				"└─Project: 1\n" +
				"  └─Filter: posts.owner_id = users.id\n" +
				"    └─Scan Table: posts [physical]\n",
		},
		{
			// tests that correlation is propagated across multiple subqueries
			name: "double nested correlated subquery",
			sql:  "select name from users u where exists (select 1 from posts p where exists (select 1 from posts p2 where p2.owner_id = u.id))",
			wt: "Return: name [text]\n" +
				"└─Project: u.name\n" +
				"  └─Filter: [subquery (exists) (subplan_id=1) (correlated: u.id)]\n" +
				"    └─Scan Table [alias=\"u\"]: users [physical]\n" +
				"Subplan [subquery] [id=1]\n" +
				"└─Project: 1\n" +
				"  └─Filter: [subquery (exists) (subplan_id=0) (correlated: u.id)]\n" +
				"    └─Scan Table [alias=\"p\"]: posts [physical]\n" +
				"Subplan [subquery] [id=0]\n" +
				"└─Project: 1\n" +
				"  └─Filter: p2.owner_id = u.id\n" +
				"    └─Scan Table [alias=\"p2\"]: posts [physical]\n",
		},
		{
			name: "aggregate without group by",
			sql:  "select sum(age) from users",
			wt: "Return: sum [decimal(1000,0)]\n" +
				"└─Project: {#ref(A)}\n" +
				"  └─Aggregate: {#ref(A) = sum(users.age)}\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "aggregate with group by",
			sql:  "select name, sum(age) from users where name = 'a' group by name having sum(age)::int > 100",
			wt: "Return: name [text], sum [decimal(1000,0)]\n" +
				"└─Project: {#ref(A)}; {#ref(B)}\n" +
				"  └─Filter: {#ref(B)}::int > 100\n" +
				"    └─Aggregate [{#ref(A) = users.name}]: {#ref(B) = sum(users.age)}\n" +
				"      └─Filter: users.name = 'a'\n" +
				"        └─Scan Table: users [physical]\n",
		},
		{
			name: "complex group by and aggregate",
			sql:  "select sum(u.age)::int/(p.created_at/100) as res from users u inner join posts p on u.id=p.owner_id group by (p.created_at/100)",
			wt: "Return: res [int]\n" +
				"└─Project: {#ref(B)}::int / {#ref(A)} AS res\n" +
				"  └─Aggregate [{#ref(A) = p.created_at / 100}]: {#ref(B) = sum(u.age)}\n" +
				"    └─Join [inner]: u.id = p.owner_id\n" +
				"      ├─Scan Table [alias=\"u\"]: users [physical]\n" +
				"      └─Scan Table [alias=\"p\"]: posts [physical]\n",
		},
		{
			name: "invalid group by column",
			sql:  "select age from users group by age/2",
			err:  logical.ErrIllegalAggregate,
		},
		{
			name: "aggregate in group by",
			sql:  "select sum(age) from users group by sum(age)",
			err:  logical.ErrIllegalAggregate,
		},
		{
			name: "aggregate in where clause",
			sql:  "select sum(age) from users where sum(age)::int > 100",
			err:  logical.ErrAggregateInWhere,
		},
		{
			name: "complex group by",
			sql:  "select age/2, age*3 from users group by age/2, age*3",
			wt: "Return: ?column? [int], ?column? [int]\n" +
				"└─Project: {#ref(A)}; {#ref(B)}\n" +
				"  └─Aggregate [{#ref(A) = users.age / 2}] [{#ref(B) = users.age * 3}]\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "select * with group by",
			sql:  "select * from users group by name, age, id",
			wt: "Return: id [uuid], name [text], age [int]\n" +
				"└─Project: {#ref(C)}; {#ref(A)}; {#ref(B)}\n" +
				"  └─Aggregate [{#ref(A) = users.name}] [{#ref(B) = users.age}] [{#ref(C) = users.id}]\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "select * with group by (not enough group by columns)",
			sql:  "select * from users group by name",
			err:  logical.ErrIllegalAggregate,
		},
		{
			name: "complex having",
			sql:  "select name, sum(age/2)+sum(age*10) from users group by name having sum(age)::int > 100 or sum(age/2)::int > 10",
			wt: "Return: name [text], ?column? [decimal(1000,0)]\n" +
				"└─Project: {#ref(A)}; {#ref(C)} + {#ref(D)}\n" +
				"  └─Filter: {#ref(B)}::int > 100 OR {#ref(C)}::int > 10\n" +
				"    └─Aggregate [{#ref(A) = users.name}]: {#ref(D) = sum(users.age * 10)}; {#ref(C) = sum(users.age / 2)}; {#ref(B) = sum(users.age)}\n" +
				"      └─Scan Table: users [physical]\n",
		},
		{
			name: "duplicate group by columns",
			sql:  "select name, age from users group by name, name, age",
			wt: "Return: name [text], age [int]\n" +
				"└─Project: {#ref(A)}; {#ref(B)}\n" +
				"  └─Aggregate [{#ref(A) = users.name}] [{#ref(B) = users.age}]\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "group by with alias",
			sql:  "select name from users group by users.name",
			wt: "Return: name [text]\n" +
				"└─Project: {#ref(A)}\n" +
				"  └─Aggregate [{#ref(A) = users.name}]\n" +
				"    └─Scan Table: users [physical]\n",
		},
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
			wt: "Return: brand [text], content [text], name [text], id [uuid], count [int]\n" +
				"└─Project: {#ref(A)}; {#ref(B)}; {#ref(C)}; {#ref(D)}; {#ref(E)}\n" +
				"  └─Aggregate [{#ref(A) = c.brand}] [{#ref(B) = pu.content}] [{#ref(C) = u.name}] [{#ref(D) = u2.id}]: {#ref(E) = count(p.id)}\n" +
				"    └─Join [outer]: u2.id = u.id\n" +
				"      ├─Join [right]: pu.content = p.content\n" +
				"      │ ├─Join [left]: c.owner_name = u.name\n" +
				"      │ │ ├─Join [inner]: u.id = p.owner_id\n" +
				"      │ │ │ ├─Scan Table [alias=\"u\"]: users [physical]\n" +
				"      │ │ │ └─Scan Table [alias=\"p\"]: posts [physical]\n" +
				"      │ │ └─Scan Procedure [alias=\"c\"]: [foreign=true] [dbid='dbid'] [proc='proc'] owned_cars($id)\n" +
				"      │ └─Scan Procedure [alias=\"pu\"]: [foreign=false] posts_by_user($name)\n" +
				"      └─Scan Subquery [alias=\"u2\"]: [subplan_id=0] (uncorrelated)\n" +
				"Subplan [subquery] [id=0]\n" +
				"└─Project: users.id\n" +
				"  └─Filter: users.age > 18\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "common table expressions",
			sql: `with a (id2, name2) as (select id, name from users),
				b as (select * from a)
				select * from b;`,
			wt: "Return: id2 [uuid], name2 [text]\n" +
				"└─Project: b.id2; b.name2\n" +
				"  └─Scan Table: b [cte]\n" +
				"Subplan [cte] [id=b] [a.id2 -> id2] [a.name2 -> name2]\n" +
				"└─Project: a.id2; a.name2\n" +
				"  └─Scan Table: a [cte]\n" +
				"Subplan [cte] [id=a] [users.id -> id2] [users.name -> name2]\n" +
				"└─Project: users.id; users.name\n" +
				"  └─Scan Table: users [physical]\n",
		},
		{
			name: "set operations",
			sql: `select id, name from users
				union
				select id, name from users
				union all
				select id, name from users
				intersect
				select id, name from users
				except
				select id, name from users;`,
			wt: "Return: id [uuid], name [text]\n" +
				"└─Set: except\n" +
				"  ├─Set: intersect\n" +
				"  │ ├─Set: union all\n" +
				"  │ │ ├─Set: union\n" +
				"  │ │ │ ├─Project: users.id; users.name\n" +
				"  │ │ │ │ └─Scan Table: users [physical]\n" +
				"  │ │ │ └─Project: users.id; users.name\n" +
				"  │ │ │   └─Scan Table: users [physical]\n" +
				"  │ │ └─Project: users.id; users.name\n" +
				"  │ │   └─Scan Table: users [physical]\n" +
				"  │ └─Project: users.id; users.name\n" +
				"  │   └─Scan Table: users [physical]\n" +
				"  └─Project: users.id; users.name\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "incompatible set schema types",
			sql: `select id, name from users
				union
				select id, owner_id from posts;`,
			err: logical.ErrSetIncompatibleSchemas,
		},
		{
			name: "incompatible set schema lengths",
			sql: `select id, name from users
				union
				select 1;`,
			err: logical.ErrSetIncompatibleSchemas,
		},
		{
			name: "set operations with order by and limit",
			sql: `select id, name from users
				union
				select id, content from posts
				order by name desc;`,
			wt: "Return: id [uuid], name [text]\n" +
				"└─Sort: [name] desc nulls last\n" +
				"  └─Set: union\n" +
				"    ├─Project: users.id; users.name\n" +
				"    │ └─Scan Table: users [physical]\n" +
				"    └─Project: posts.id; posts.content\n" +
				"      └─Scan Table: posts [physical]\n",
		},
		{
			name: "sort",
			sql:  "select name, age from users order by name desc nulls last, id asc",
			wt: "Return: name [text], age [int]\n" +
				"└─Project: users.name; users.age\n" +
				"  └─Sort: [users.name] desc nulls last; [users.id] asc nulls last\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "sort invalid column",
			sql:  "select name, age from users order by wallet",
			err:  logical.ErrColumnNotFound,
		},
		{
			name: "limit and offset",
			sql:  "select name, age from users limit 10 offset 5",
			wt: "Return: name [text], age [int]\n" +
				"└─Project: users.name; users.age\n" +
				"  └─Limit [offset=5]: 10\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "distinct",
			sql:  "select distinct name, age from users",
			wt: "Return: name [text], age [int]\n" +
				"└─Distinct\n" +
				"  └─Project: users.name; users.age\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "scalar function, procedure, foreign procedure",
			sql:  "select car_count['dbid', 'proc'](id), post_count(id), abs(age) from users",
			wt: "Return: car_count [int], post_count [int], abs [int]\n" +
				"└─Project: car_count['dbid', 'proc'](users.id); post_count(users.id); abs(users.age)\n" +
				"  └─Scan Table: users [physical]\n",
		},
		{
			name: "distinct aggregate",
			sql:  "select count(distinct name), sum(age) from users",
			wt: "Return: count [int], sum [decimal(1000,0)]\n" +
				"└─Project: {#ref(A)}; {#ref(B)}\n" +
				"  └─Aggregate: {#ref(A) = count(distinct users.name)}; {#ref(B) = sum(users.age)}\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "unary and alias",
			sql:  "select age as pos_age, -age from users",
			wt: "Return: pos_age [int], ?column? [int]\n" +
				"└─Project: users.age AS pos_age; -users.age\n" +
				"  └─Scan Table: users [physical]\n",
		},
		{
			name: "order by alias",
			sql:  "select age as pos_age from users order by pos_age",
			wt: "Return: pos_age [int]\n" +
				"└─Project: users.age AS pos_age\n" +
				"  └─Sort: [pos_age] asc nulls last\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "collate",
			sql:  "select name collate nocase from users where name = 'SATOSHI' collate nocase",
			wt: "Return: name [text]\n" +
				"└─Project: users.name COLLATE nocase\n" +
				"  └─Filter: users.name = 'SATOSHI' COLLATE nocase\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "in",
			sql:  "select name from users where name not in ('satoshi', 'wendys_drive_through_lady')",
			wt: "Return: name [text]\n" +
				"└─Project: users.name\n" +
				// planner rewrites NOT IN to a unary NOT(IN) for simplicity, since it's equivalent
				"  └─Filter: NOT users.name IN ('satoshi', 'wendys_drive_through_lady')\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "like and ilike",
			// planner rewrites NOT LIKE/ILIKE to unary NOT(LIKE/ILIKE) for simplicity
			sql: "select name from users where name like 's%' or name not ilike 'w_Nd%'",
			wt: "Return: name [text]\n" +
				"└─Project: users.name\n" +
				"  └─Filter: users.name LIKE 's%' OR NOT users.name ILIKE 'w_Nd%'\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "case",
			sql:  `select name from users where case age when 20 then true else false end`,
			wt: "Return: name [text]\n" +
				"└─Project: users.name\n" +
				"  └─Filter: CASE [users.age] WHEN [20] THEN [true] ELSE [false] END\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "case (no expression)",
			sql:  `select name from users where case when age = 20 then true else false end`,
			wt: "Return: name [text]\n" +
				"└─Project: users.name\n" +
				"  └─Filter: CASE WHEN [users.age = 20] THEN [true] ELSE [false] END\n" +
				"    └─Scan Table: users [physical]\n",
		},
		{
			name: "basic update",
			sql:  "update users set name = 'satoshi' where age = 1",
			wt: "Update [users]: name = 'satoshi'\n" +
				"└─Filter: users.age = 1\n" +
				"  └─Scan Table: users [physical]\n",
		},
		{
			name: "update from with join",
			sql:  "update users set name = pu.content from posts p inner join posts_by_user('satoshi') pu on p.content = pu.content where p.owner_id = users.id",
			// will be unoptimized, so it will use a cartesian product
			// optimization could re-write the filter to a join, as well as
			// add projections.
			wt: "Update [users]: name = pu.content\n" +
				"└─Filter: p.owner_id = users.id\n" +
				"  └─Cartesian Product\n" +
				"    ├─Scan Table: users [physical]\n" +
				"    └─Join [inner]: p.content = pu.content\n" +
				"      ├─Scan Table [alias=\"p\"]: posts [physical]\n" +
				"      └─Scan Procedure [alias=\"pu\"]: [foreign=false] posts_by_user('satoshi')\n",
		},
		{
			name: "update with from without where",
			sql:  "update users set name = pu.content from posts_by_user('satoshi') pu",
			err:  logical.ErrUpdateOrDeleteWithoutWhere,
		},
		{
			name: "basic delete",
			sql:  "delete from users where age = 1",
			wt: "Delete [users]\n" +
				"└─Filter: users.age = 1\n" +
				"  └─Scan Table: users [physical]\n",
		},
		{
			name: "insert",
			sql:  "insert into users values ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', 1), ('123e4567-e89b-12d3-a456-426614174001'::uuid, 'satoshi2', 2)",
			wt: "Insert [users]: id [uuid], name [text], age [int]\n" +
				"└─Values: ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', 1); ('123e4567-e89b-12d3-a456-426614174001'::uuid, 'satoshi2', 2)\n",
		},
		{
			name: "insert with null",
			sql:  "insert into users (id, name) values ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi')",
			wt: "Insert [users]: id [uuid], name [text], age [int]\n" +
				"└─Values: ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', NULL)\n",
		},
		{
			name: "insert null in non-nullable column",
			sql:  "insert into users (name) values ('satoshi')",
			err:  logical.ErrNotNullableColumn,
		},
		{
			name: "on conflict do nothing",
			sql:  "insert into users values ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', 1) on conflict do nothing",
			wt: "Insert [users]: id [uuid], name [text], age [int]\n" +
				"├─Values: ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', 1)\n" +
				"└─Conflict [nothing]\n",
		},
		{
			name: "on conflict do update (arbiter index primary key)",
			sql:  "insert into users values ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', 1) on conflict (id) do update set name = 'satoshi'",
			wt: "Insert [users]: id [uuid], name [text], age [int]\n" +
				"├─Values: ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', 1)\n" +
				"└─Conflict [update] [arbiter=users.id (primary key)]: [name = 'satoshi']\n",
		},
		{
			name: "on conflict do update (arbiter unique constraint)",
			sql:  "insert into posts values ('123e4567-e89b-12d3-a456-426614174000'::uuid, '123e4567-e89b-12d3-a456-426614174001'::uuid, 'hello', 1) on conflict (content) do update set owner_id = '123e4567-e89b-12d3-a456-426614174001'::uuid",
			wt: "Insert [posts]: id [uuid], owner_id [uuid], content [text], created_at [int]\n" +
				"├─Values: ('123e4567-e89b-12d3-a456-426614174000'::uuid, '123e4567-e89b-12d3-a456-426614174001'::uuid, 'hello', 1)\n" +
				"└─Conflict [update] [arbiter=posts.content (unique)]: [owner_id = '123e4567-e89b-12d3-a456-426614174001'::uuid]\n",
		},
		{
			name: "on conflict do update (arbiter index non-primary key)",
			sql:  "insert into users values ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', 1) on conflict (name) do update set name = 'satoshi' WHERE users.age = 1",
			wt: "Insert [users]: id [uuid], name [text], age [int]\n" +
				"├─Values: ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', 1)\n" +
				"└─Conflict [update] [arbiter=name_idx (index)]: [name = 'satoshi'] where [users.age = 1]\n",
		},
		{
			name: "on conflict with non-arbiter column",
			sql:  "insert into users values ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', 1) on conflict (age) do update set name = 'satoshi'",
			err:  logical.ErrIllegalConflictArbiter,
		},
		{
			name: "on conflict with half of a composite index",
			sql:  "insert into posts values ('123e4567-e89b-12d3-a456-426614174000'::uuid, '123e4567-e89b-12d3-a456-426614174001'::uuid, 'hello', 1) on conflict (owner_id) do update set content = 'hello'",
			err:  logical.ErrIllegalConflictArbiter,
		},
		{
			name: "on conflict with non-unique index",
			sql:  "insert into users values ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', 1) on conflict (age) do update set name = 'satoshi'",
			err:  logical.ErrIllegalConflictArbiter,
		},
		{
			name: "conflict on compound unique index",
			sql:  "insert into posts values ('123e4567-e89b-12d3-a456-426614174000'::uuid, '123e4567-e89b-12d3-a456-426614174001'::uuid, 'hello', 1) on conflict (owner_id, created_at) do update set content = 'hello'",
			wt: "Insert [posts]: id [uuid], owner_id [uuid], content [text], created_at [int]\n" +
				"├─Values: ('123e4567-e89b-12d3-a456-426614174000'::uuid, '123e4567-e89b-12d3-a456-426614174001'::uuid, 'hello', 1)\n" +
				"└─Conflict [update] [arbiter=owner_created_idx (index)]: [content = 'hello']\n",
		},
		{
			name: "excluded clause",
			sql:  "insert into users (id, name) values ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi') on conflict (id) do update set name = excluded.name where (excluded.age/2) = 0",
			wt: "Insert [users]: id [uuid], name [text], age [int]\n" +
				"├─Values: ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', NULL)\n" +
				"└─Conflict [update] [arbiter=users.id (primary key)]: [name = excluded.name] where [excluded.age / 2 = 0]\n",
		},
		{
			// surprisingly, this mirrors Postgres's behavior
			name: "ambiguous column due to excluded",
			sql:  "insert into users values ('123e4567-e89b-12d3-a456-426614174000'::uuid, 'satoshi', 1) on conflict (name) do update set name = 'satoshi' WHERE age = 1",
			err:  logical.ErrAmbiguousColumn,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := parse.Parse([]byte(testSchema))
			require.NoError(t, err)

			parsedSql, err := parse.ParseSQL(test.sql, schema, true)
			require.NoError(t, err)
			require.NoError(t, parsedSql.ParseErrs.Err())

			plan, err := logical.CreateLogicalPlan(parsedSql.AST, schema, test.vars, test.objects)
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

				// check that Relation() works
				plan.Plan.Relation()

				for _, cte := range plan.CTEs {
					cte.Relation()
				}

				// make sure nothing changed
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
