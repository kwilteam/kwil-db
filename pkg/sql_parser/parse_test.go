package sql_parser

import (
	"flag"
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/tree"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var trace = flag.Bool("trace", false, "run tests with tracing")

type selectTestCases struct {
	name   string
	input  string
	expect tree.Ast
}

//func genSelectLiteralCases() []*tree.Select {
//	ps := []selectTestCases{
//		{"number", "select 1", nil},
//		{"string", "select 'a'", nil},
//		{"null", "select null", nil},
//		{"true", "select true", nil},
//		{"false", "select false", nil},
//		{"blob", "select x'01'", nil},
//	}
//
//	base := tree.Select{
//		SelectStmt: &tree.SelectStmt{
//			SelectCores: []*tree.SelectCore{},
//		},
//	}
//
//	return
//}

func TestParseRawSQL_visitor_allowed(t *testing.T) {

	tests := []struct {
		name   string
		input  string
		expect tree.Ast
	}{
		//// literal value
		{"number", "select 1",
			&tree.Select{
				SelectStmt: &tree.SelectStmt{
					SelectCores: []*tree.SelectCore{
						{
							SelectType: tree.SelectTypeAll,
							Columns:    []string{"1"},
						},
					},
				}}},
		//{"string", "select 'a'"},
		//{"null", "select null"},
		//{"true", "select true"},
		//{"false", "select false"},
		//{"blob", "select x'01'"},
		//
		//// common table stmt
		//{"cte", "with t as (select 1) select * from t"},
		//{"cte with column", "with t1(c1,c2) as (select 1) select * from t"},
		//
		//// compound operator
		//{"union", "select 1 union select 2"},
		//{"union all", "select 1 union all select 2"},
		//{"intersect", "select 1 intersect select 2"},
		//{"except", "select 1 except select 2"},
		//
		//// table or subquery
		//{"table or subquery", "select * from t1 as tt"},
		//{"table or subquery nest select", "select * from (select 1) as tt"},
		//{"table or subquery nest tos", "select * from (t1 as tt, t2 as ttt)"},
		//{"table or subquery join", "select * from t1 as tt join t2 as ttt on tt.a = ttt.a"},
		//
		//// expr
		//{"expr bind parameter ?", "select ?"},
		//{"expr bind parameter $", "select $a"},
		//{"expr bind parameter @", "select @a"},
		//{"expr bind parameter :", "select :a"},
		//{"expr names", "select schema.table.column"},
		////
		//{"expr unary op +", "select +1"},
		//{"expr unary op -", "select -1"},
		//{"expr unary op ~", "select ~1"},
		////
		//{"expr unary op not", "select not 1"},
		//{"expr binary op ||", "select 1 || 2"},
		//{"expr binary op *", "select 1 * 2"},
		//{"expr binary op /", "select 1 / 2"},
		//{"expr binary op %", "select 1 % 2"},
		//{"expr binary op +", "select 1 + 2"},
		//{"expr binary op -", "select 1 - 2"},
		//{"expr binary op <<", "select 1 << 2"},
		//{"expr binary op >>", "select 1 >> 2"},
		//{"expr binary op &", "select 1 & 2"},
		//{"expr binary op |", "select 1 | 2"},
		//{"expr binary op <", "select 1 < 2"},
		//{"expr binary op <=", "select 1 <= 2"},
		//{"expr binary op >", "select 1 > 2"},
		//{"expr binary op >=", "select 1 >= 2"},
		//{"expr binary op =", "select 1 = 2"},
		//{"expr binary op !=", "select 1 != 2"},
		//{"expr binary op <>", "select 1 <> 2"},
		//{"expr binary op is", "select 1 is 2"},
		//{"expr binary op is not", "select 1 is not 2"},
		//{"expr binary op in", "select 1 in (1,2)"},
		//{"expr binary op not in", "select 1 not in (1,2)"},
		//{"expr binary op like", "select 1 like 2"},
		//{"expr binary op match", "select 1 match 2"},
		//{"expr binary op regexp", "select 1 regexp 2"},
		//{"expr binary op and", "select 1 and 2"},
		//{"expr binary op or", "select 1 or 2"},
		////
		//{"expr function no param", "select f()"},
		//{"expr function one param", "select f(1)"},
		//{"expr function multi param", "select f(1,2)"},
		//{"expr function distinct param", "select f(distinct 1,2)"},
		//{"expr function * param", "select f(*)"},
		//{"expr function with filter", "select f(1) filter (where 1)"},
		////
		//{"expr in parentheses", "select (1)"},
		////
		//{"expr with collate", "select 1 collate nocase"},
		////
		//{"expr like", "select 1 like 2"},
		//{"expr like escape", "select 1 like 2 escape 3"},
		//{"expr not like", "select 1 not like 2"},
		//{"expr match", "select 1 match 2"},
		//{"expr not match", "select 1 not match 2"},
		//{"expr regexp", "select 1 regexp 2"},
		//{"expr not regexp", "select 1 not regexp 2"},
		//// null
		//{"expr isnull", "select 1 isnull"},
		//{"expr notnull", "select 1 notnull"},
		//{"expr not null", "select 1 not null"},
		////
		//{"expr is", "select 1 is 2"},
		//{"expr is not", "select 1 is not 2"},
		//{"expr is distinct from", "select 1 is not distinct from 2"},
		////
		//{"expr between", "select 1 between 2 and 3"},
		//{"expr not between", "select 1 not between 2 and 3"},
		////
		//{"expr in", "select 1 in (1,2)"},
		//{"expr not in", "select 1 not in (1,2)"},
		//{"expr in subquery", "select 1 in (select 1)"},
		////
		//{"expr exists", "select exists (select 1)"},
		//{"expr not exists", "select not exists (select 1)"},
		////
		//{"expr case", "select case when 1 then 2 end"},
		//{"expr case else", "select case when 1 then 2 else 3 end"},
		//{"expr case multi when", "select case when 1 then 2 when 3 then 4 end"},
		//{"expr case expr", "select case 1 when 2 then 3 end"},

		// insert stmt
		//{"insert", "insert into t1 values ('1')", tree.Insert{
		//	InsertStmt: &tree.InsertStmt{
		//		Table:      "t1",
		//		InsertType: tree.InsertTypeInsert,
		//		Values:     [][]tree.Expression{{&tree.ExpressionLiteral{Value: "1"}}},
		//	},
		//}},
		{"insert", "insert into t1 values (1)",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					InsertType: tree.InsertTypeInsert,
					Values:     [][]tree.Expression{{&tree.ExpressionLiteral{Value: "1"}}},
				}}},
		{"insert replace", "replace into t1 values (1)",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					InsertType: tree.InsertTypeReplace,
					Values:     [][]tree.Expression{{&tree.ExpressionLiteral{Value: "1"}}},
				}}},
		{"insert or replace", "insert or replace into t1 values (1)",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					InsertType: tree.InsertTypeInsertOrReplace,
					Values:     [][]tree.Expression{{&tree.ExpressionLiteral{Value: "1"}}},
				}}},
		{"insert with columns", "insert into t1 (a,b) values (1,2)",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Values: [][]tree.Expression{
						{
							&tree.ExpressionLiteral{Value: "1"},
							&tree.ExpressionLiteral{Value: "2"},
						},
					}}}},
		//{"insert with cte", "with t as (select 1) insert into t1 (a,b) values (1,2)",
		//	&tree.Insert{
		//		CTE: []*tree.CTE{
		//			{
		//				Table: "t",
		//				Select: &tree.SelectStmt{
		//					SelectCore: nil,
		//					OrderBy:    nil,
		//					Limit:      nil,
		//				},
		//			},
		//		},
		//		InsertStmt: &tree.InsertStmt{
		//			Table:      "t1",
		//			Columns:    []string{"a", "b"},
		//			InsertType: tree.InsertTypeInsert,
		//			Values: [][]tree.Expression{
		//				{&tree.ExpressionLiteral{Value: "1"}, &tree.ExpressionLiteral{Value: "2"}},
		//			}}}},
		//{"insert with returning", "insert into t1 (a,b) values (1,2) returning a",
		//	&tree.Insert{
		//		CTE: nil,
		//		InsertStmt: &tree.InsertStmt{
		//			Table:      "t1",
		//			Columns:    []string{"a", "b"},
		//			InsertType: tree.InsertTypeInsert,
		//			Values: [][]tree.Expression{
		//				{&tree.ExpressionLiteral{Value: "1"}, &tree.ExpressionLiteral{Value: "2"}},
		//			},
		//			//ReturningClause:
		//		}}},
		//{"insert with returning *", "insert into t1 (a,b) values (1,2) returning *"},
		//{"insert with or replace", "insert or replace into t1 (a,b) values (1,2)"},
		//{"insert with table alias", "insert into t1 as t (a,b) values (1,2)"},

		// TODO: @yaiba generate expected tree
		{"insert with values upsert without target do nothing", "insert into t1 (a,b) values (1,2) on conflict do nothing",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoNothing,
					},
					Values: [][]tree.Expression{
						{&tree.ExpressionLiteral{Value: "1"}, &tree.ExpressionLiteral{Value: "2"}},
					}}}},
		{"insert with values upsert with target without where do nothing",
			"insert into t1 (a,b) values (1,2) on conflict (c1,c2) do nothing",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						ConflictTarget: &tree.ConflictTarget{
							IndexedColumns: []string{"c1", "c2"},
						},
						Type: tree.UpsertTypeDoNothing,
					},
					Values: [][]tree.Expression{
						{&tree.ExpressionLiteral{Value: "1"}, &tree.ExpressionLiteral{Value: "2"}},
					}}}},
		{"insert with values upsert with target and where do nothing",
			"insert into t1 (a,b) values (1,2) on conflict(c1,c2) where 1 do nothing",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						ConflictTarget: &tree.ConflictTarget{
							IndexedColumns: []string{"c1", "c2"},
							Where:          &tree.ExpressionLiteral{"1"},
						},
						Type: tree.UpsertTypeDoNothing,
					},
					Values: [][]tree.Expression{
						{&tree.ExpressionLiteral{Value: "1"}, &tree.ExpressionLiteral{Value: "2"}},
					}}}},
		{"insert with values upsert with update column name",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoUpdate,
						Updates: []*tree.UpdateSetClause{
							{Columns: []string{"b"},
								Expression: &tree.ExpressionLiteral{Value: "1"}},
						},
					},
					Values: [][]tree.Expression{
						{&tree.ExpressionLiteral{Value: "1"}, &tree.ExpressionLiteral{Value: "2"}},
					}}}},
		{"insert with values upsert with update multi column name",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1,c=2",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoUpdate,
						Updates: []*tree.UpdateSetClause{
							{
								Columns:    []string{"b"},
								Expression: &tree.ExpressionLiteral{Value: "1"},
							},
							{
								Columns:    []string{"c"},
								Expression: &tree.ExpressionLiteral{Value: "2"},
							},
						},
					},
					Values: [][]tree.Expression{
						{&tree.ExpressionLiteral{Value: "1"}, &tree.ExpressionLiteral{Value: "2"}},
					}}}},
		{"insert with values upsert with update column name list",
			"insert into t1 (a,b) values (1,2) on conflict do update set (b,c)=(1,2)",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoUpdate,
						Updates: []*tree.UpdateSetClause{
							{
								Columns: []string{"b", "c"},
								Expression: &tree.ExpressionList{
									Expressions: []tree.Expression{
										&tree.ExpressionLiteral{Value: "1"},
										&tree.ExpressionLiteral{Value: "2"},
									}},
							},
						},
					},
					Values: [][]tree.Expression{
						{&tree.ExpressionLiteral{Value: "1"}, &tree.ExpressionLiteral{Value: "2"}},
					}}}},
		{"insert with values upsert with update multi column name list",
			"insert into t1 (a,b) values (1,2) on conflict do update set (b,c)=(1,2), (d,e)=(3,4)",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoUpdate,
						Updates: []*tree.UpdateSetClause{
							{
								Columns: []string{"b", "c"},
								Expression: &tree.ExpressionList{
									Expressions: []tree.Expression{
										&tree.ExpressionLiteral{Value: "1"},
										&tree.ExpressionLiteral{Value: "2"},
									}},
							},
							{
								Columns: []string{"d", "e"},
								Expression: &tree.ExpressionList{
									Expressions: []tree.Expression{
										&tree.ExpressionLiteral{Value: "3"},
										&tree.ExpressionLiteral{Value: "4"},
									}},
							},
						},
					},
					Values: [][]tree.Expression{
						{&tree.ExpressionLiteral{Value: "1"}, &tree.ExpressionLiteral{Value: "2"}},
					}}}},
		{"insert with values upsert with update and where",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1 where 1",
			&tree.Insert{
				CTE: nil,
				InsertStmt: &tree.InsertStmt{
					Table:      "t1",
					Columns:    []string{"a", "b"},
					InsertType: tree.InsertTypeInsert,
					Upsert: &tree.Upsert{
						Type: tree.UpsertTypeDoUpdate,
						Updates: []*tree.UpdateSetClause{
							{Columns: []string{"b"},
								Expression: &tree.ExpressionLiteral{Value: "1"}},
						},
						Where: &tree.ExpressionLiteral{Value: "1"},
					},
					Values: [][]tree.Expression{
						{&tree.ExpressionLiteral{Value: "1"}, &tree.ExpressionLiteral{Value: "2"}},
					}}}},

		//// join
		//{"join on", "select * from t1 join t2 on t1.c1=t2.c1"},
		//{"join implicit", "select * from t1,t2 on t1.c1=t2.c1"},
		//{"left join", "select * from t1 left join t2 on t1.c1=t2.c1"},
		//{"left outer join", "select * from t1 left outer join t2 on t1.c1=t2.c1"},
		//{"right join", "select * from t1 right join t2 on t1.c1=t2.c1"},
		//{"right outer join", "select * from t1 right outer join t2 on t1.c1=t2.c1"},
		//{"full join", "select * from t1 full join t2 on t1.c1=t2.c1"},
		//{"full outer join", "select * from t1 full outer join t2 on t1.c1=t2.c1"},
		//{"inner join", "select * from t1 inner join t2 on t1.c1=t2.c1"},
		//
		//// select
		//{"select *", "select * from t1"},
		//{"select with cte", "with t as (select 1) select * from t1"},
		//{"select distinct", "select distinct * from t1"},
		//{"select from join clause", "select * from t1 join t2 on t1.c1=t2.c1"},
		//{"select with where", "select * from t1 where 1"},
		//{"select with group by", "select * from t1 group by c1"},
		//{"select with group by and having", "select * from t1 group by c1 having 1"},
		//{"select values", "values (1)"},
		//{"select values with cte", "with t as (select 1) values (1)"},
		//{"select with compound operator union", "select * from t1 union select * from t2"},
		//{"select with compound operator union all", "select * from t1 union all select * from t2"},
		//{"select with compound operator intersect", "select * from t1 intersect select * from t2"},
		//{"select with compound operator except", "select * from t1 except select * from t2"},
		//{"select with compound operator and values", "select * from t1 union values (1)"},
		//{"select with order by", "select * from t1 order by c1 collate collate_name asc"},
		//{"select with limit", "select * from t1 limit 1"},
		//{"select with limit offset", "select * from t1 limit 1 offset 2"},
		//{"select with limit comma", "select * from t1 limit 1,10"},
	}

	ctx := DatabaseContext{Actions: map[string]ActionContext{"action1": {}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := NewErrorHandler(1)
			el := newSqliteErrorListener(eh)
			ast, err := ParseRawSQLVisitor(tt.input, 1, "action1", ctx, el, *trace, false)
			if err != nil {
				t.Errorf("ParseRawSQL() got %s", err)
			}

			astNodes := ast.(asts)
			node := astNodes[0]
			//fmt.Printf("AST: %+v\n", node.(*tree.Insert).InsertStmt)
			//fmt.Printf("exp: %+v\n", tt.expect.(*tree.Insert).InsertStmt)
			// use assert.Exactly?
			assert.EqualValues(t, tt.expect, node, "ParseRawSQL() got %s, want %s", node, tt.expect)

			fmt.Printf("AST: %+v\n", node)
			_, err = node.(tree.Ast).ToSQL()
			if err != nil {
				t.Errorf("ParseRawSQL() got %s", err)
			}
			//
			//if sql != tt.input {
			//	t.Errorf("ParseRawSQL() got %s, want %s", sql, tt.input)
			//}
		})
	}
}

func TestParseRawSQL_listener_allowed(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// literal value
		{"number", "select 1"},
		{"string", "select 'a'"},
		{"null", "select null"},
		{"true", "select true"},
		{"false", "select false"},
		{"blob", "select x'01'"},

		// common table stmt
		{"cte", "with t as (select 1) select * from t"},
		{"cte with column", "with t1(c1,c2) as (select 1) select * from t"},

		// compound operator
		{"union", "select 1 union select 2"},
		{"union all", "select 1 union all select 2"},
		{"intersect", "select 1 intersect select 2"},
		{"except", "select 1 except select 2"},

		// table or subquery
		{"table or subquery", "select * from t1 as tt"},
		{"table or subquery nest select", "select * from (select 1) as tt"},
		{"table or subquery nest tos", "select * from (t1 as tt, t2 as ttt)"},
		{"table or subquery join", "select * from t1 as tt join t2 as ttt on tt.a = ttt.a"},

		// expr
		{"expr bind parameter ?", "select ?"},
		{"expr bind parameter $", "select $a"},
		{"expr bind parameter @", "select @a"},
		{"expr bind parameter :", "select :a"},
		{"expr names", "select schema.table.column"},
		//
		{"expr unary op +", "select +1"},
		{"expr unary op -", "select -1"},
		{"expr unary op ~", "select ~1"},
		//
		{"expr unary op not", "select not 1"},
		{"expr binary op ||", "select 1 || 2"},
		{"expr binary op *", "select 1 * 2"},
		{"expr binary op /", "select 1 / 2"},
		{"expr binary op %", "select 1 % 2"},
		{"expr binary op +", "select 1 + 2"},
		{"expr binary op -", "select 1 - 2"},
		{"expr binary op <<", "select 1 << 2"},
		{"expr binary op >>", "select 1 >> 2"},
		{"expr binary op &", "select 1 & 2"},
		{"expr binary op |", "select 1 | 2"},
		{"expr binary op <", "select 1 < 2"},
		{"expr binary op <=", "select 1 <= 2"},
		{"expr binary op >", "select 1 > 2"},
		{"expr binary op >=", "select 1 >= 2"},
		{"expr binary op =", "select 1 = 2"},
		{"expr binary op !=", "select 1 != 2"},
		{"expr binary op <>", "select 1 <> 2"},
		{"expr binary op is", "select 1 is 2"},
		{"expr binary op is not", "select 1 is not 2"},
		{"expr binary op in", "select 1 in (1,2)"},
		{"expr binary op not in", "select 1 not in (1,2)"},
		{"expr binary op like", "select 1 like 2"},
		{"expr binary op match", "select 1 match 2"},
		{"expr binary op regexp", "select 1 regexp 2"},
		{"expr binary op and", "select 1 and 2"},
		{"expr binary op or", "select 1 or 2"},
		//
		{"expr function no param", "select f()"},
		{"expr function one param", "select f(1)"},
		{"expr function multi param", "select f(1,2)"},
		{"expr function distinct param", "select f(distinct 1,2)"},
		{"expr function * param", "select f(*)"},
		{"expr function with filter", "select f(1) filter (where 1)"},
		//
		{"expr in parentheses", "select (1)"},
		//
		{"expr with collate", "select 1 collate nocase"},
		//
		{"expr like", "select 1 like 2"},
		{"expr like escape", "select 1 like 2 escape 3"},
		{"expr not like", "select 1 not like 2"},
		{"expr match", "select 1 match 2"},
		{"expr not match", "select 1 not match 2"},
		{"expr regexp", "select 1 regexp 2"},
		{"expr not regexp", "select 1 not regexp 2"},
		// null
		{"expr isnull", "select 1 isnull"},
		{"expr notnull", "select 1 notnull"},
		{"expr not null", "select 1 not null"},
		//
		{"expr is", "select 1 is 2"},
		{"expr is not", "select 1 is not 2"},
		{"expr is distinct from", "select 1 is not distinct from 2"},
		//
		{"expr between", "select 1 between 2 and 3"},
		{"expr not between", "select 1 not between 2 and 3"},
		//
		{"expr in", "select 1 in (1,2)"},
		{"expr not in", "select 1 not in (1,2)"},
		{"expr in subquery", "select 1 in (select 1)"},
		//
		{"expr exists", "select exists (select 1)"},
		{"expr not exists", "select not exists (select 1)"},
		//
		{"expr case", "select case when 1 then 2 end"},
		{"expr case else", "select case when 1 then 2 else 3 end"},
		{"expr case multi when", "select case when 1 then 2 when 3 then 4 end"},
		{"expr case expr", "select case 1 when 2 then 3 end"},

		{"insert", "insert into t1 values ('1')"},
		{"insert", "insert into t1 values (1)"},
		{"insert replace", "replace into t1 values (1)"},
		{"insert or replace", "insert or replace into t1 values (1)"},
		{"insert with columns", "insert into t1 (a,b) values (1,2)"},
		{"insert with cte", "with t as (select 1) insert into t1 (a,b) values (1,2)"},
		{"insert with returning", "insert into t1 (a,b) values (1,2) returning a"},
		{"insert with returning *", "insert into t1 (a,b) values (1,2) returning *"},
		{"insert with or replace", "insert or replace into t1 (a,b) values (1,2)"},
		{"insert with table alias", "insert into t1 as t (a,b) values (1,2)"},

		{"insert with values upsert without target do nothing", "insert into t1 (a,b) values (1,2) on conflict do nothing"},
		{"insert with values upsert with target without where do nothing",
			"insert into t1 (a,b) values (1,2) on conflict (c1,c2) do nothing"},
		{"insert with values upsert with target and where do nothing",
			"insert into t1 (a,b) values (1,2) on conflict(c1,c2) where 1 do nothing"},
		{"insert with values upsert with update column name",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1"},
		{"insert with values upsert with update multi column name",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1,c=2"},
		{"insert with values upsert with update column name list",
			"insert into t1 (a,b) values (1,2) on conflict do update set (b,c)=(1,2)"},
		{"insert with values upsert with update multi column name list",
			"insert into t1 (a,b) values (1,2) on conflict do update set (b,c)=(1,2), (d,e)=(3,4)"},
		{"insert with values upsert with update and where",
			"insert into t1 (a,b) values (1,2) on conflict do update set b=1 where 1"},

		// join
		{"join on", "select * from t1 join t2 on t1.c1=t2.c1"},
		{"join implicit", "select * from t1,t2 on t1.c1=t2.c1"},
		{"left join", "select * from t1 left join t2 on t1.c1=t2.c1"},
		{"left outer join", "select * from t1 left outer join t2 on t1.c1=t2.c1"},
		{"right join", "select * from t1 right join t2 on t1.c1=t2.c1"},
		{"right outer join", "select * from t1 right outer join t2 on t1.c1=t2.c1"},
		{"full join", "select * from t1 full join t2 on t1.c1=t2.c1"},
		{"full outer join", "select * from t1 full outer join t2 on t1.c1=t2.c1"},
		{"inner join", "select * from t1 inner join t2 on t1.c1=t2.c1"},

		// select
		{"select *", "select * from t1"},
		{"select with cte", "with t as (select 1) select * from t1"},
		{"select distinct", "select distinct * from t1"},
		{"select from join clause", "select * from t1 join t2 on t1.c1=t2.c1"},
		{"select with where", "select * from t1 where 1"},
		{"select with group by", "select * from t1 group by c1"},
		{"select with group by and having", "select * from t1 group by c1 having 1"},
		{"select values", "values (1)"},
		{"select values with cte", "with t as (select 1) values (1)"},
		{"select with compound operator union", "select * from t1 union select * from t2"},
		{"select with compound operator union all", "select * from t1 union all select * from t2"},
		{"select with compound operator intersect", "select * from t1 intersect select * from t2"},
		{"select with compound operator except", "select * from t1 except select * from t2"},
		{"select with compound operator and values", "select * from t1 union values (1)"},
		{"select with order by", "select * from t1 order by c1 collate collate_name asc"},
		{"select with limit", "select * from t1 limit 1"},
		{"select with limit offset", "select * from t1 limit 1 offset 2"},
		{"select with limit comma", "select * from t1 limit 1,10"},
	}

	ctx := DatabaseContext{Actions: map[string]ActionContext{"action1": {}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := NewErrorHandler(1)
			el := newSqliteErrorListener(eh)
			err := ParseRawSQL(tt.input, 1, "action1", ctx, el, *trace, false)
			if err != nil {
				t.Errorf("ParseRawSQL() got %s", err)
			}
		})
	}
}

func TestParseRawSQL_syntax_not_allowed(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		causeSymbol string
	}{
		// literal value
		{"current_date", "select current_date", "current_date"},
		{"current_time", "select current_time", "current_time"},
		{"current_timestamp", "select current_timestamp", "current_timestamp"},

		// common table stmt
		{"cte recursive", "with recursive t1(c1,c2) as (select 1) select * from t1", "recursive"},
		// common table expression
		{"cte not", "with t1(c1,c2) as (select 1) not select * from t1", "not"},
		{"cte materialized", "with t1(c1,c2) as (select 1) materialized select * from t1", "materialized"},

		// table or subquery
		{"table or subquery indexed", "select * from t1 indexed by index_a", "indexed"},
		{"table or subquery not indexed", "select * from t1 not indexed", "not"},
		// NOTE: what is table function??
		{"table or subquery table function", "SELECT value FROM f(1)", "("},

		// expr
		{"expr cast", "select cast(true as aaa)", "cast"},
		{"expr binary op glob", "select 1 glob 1", "glob"},
		{"expr binary op not glob", "select 1 not glob 1", "not"},
		{"expr function with over", "select abs(1) over (partition by 1)", "over"},
		//{"expr raise", "select raise(fail, 'dsd')", "raise"},

		// insert
		{"insert or abort", "insert or abort into t1 values (1)", "abort"},
		{"insert or fail", "insert or fail into t1 values (1)", "fail"},
		{"insert or ignore", "insert or ignore into t1 values (1)", "ignore"},
		{"insert or rollback", "insert or rollback into t1 values (1)", "rollback"},
		{"insert schema_name", "insert or replace into schema.t1 values (1)", "."},
		{"insert into with select", "insert into t1 as tt with t1 as (select 1) select * from t2", "with"},
		{"insert into select", "insert into t1 as tt select * from t2", "select"},
		{"insert into default values", "insert into t1 default values", "default"},
		//wrong indexed_column syntax
		//"insert into t1 (a,b) values (1,2) on conflict (c1 collate collate_name asc) do nothing",

		// select
		{"select all", "select all c1 from t1", "c1"},
		{"select with window", "select * from t1 window w1 as (partition by 1)", "window"},

		// join
		{"natural join", "select * from t3 natural join t4", "natural"},
		{"cross join", "select * from t3 cross join t4", "cross"},
		{"join using", "select * from t3 join t4 using (c1)", "using"},
		{"join without condition", "select * from t3 join t4", "<EOF>"},
	}

	ctx := DatabaseContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := NewErrorHandler(1)
			el := newSqliteErrorListener(eh)
			err := ParseRawSQL(tt.input, 1, "action1", ctx, el, *trace, false)

			if err == nil || !strings.Contains(err.Error(), ErrSyntax.Error()) {
				t.Errorf("ParseRawSQL() expected error: %s, got %s", ErrSyntax, err)
			}

			if el.symbol != tt.causeSymbol {
				t.Errorf("ParseRawSQL() expected cause symbol: %s, got: %s", tt.causeSymbol, el.symbol)
			}
		})
	}
}

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
		// non-exist table/column
		{"select from table", "select * from t1", nil},
		{"select with CTE", "with tt as (select * from t1) select * from tt", nil},
		{"select non-exist table", "select * from t10", ErrTableNotFound},
		// joins
		{"joins 1", "select * from t1 join t2 on (1+2)=2", ErrJoinConditionTooDeep},
		{"joins 2", "select * from t1 join t2 on t1.c1=t7.c1 ", ErrTableNotFound},
		{"joins 3", "select * from t1 join t2 on t1.c1=t2.c7", ErrColumnNotFound},

		{"multi joins", `select * from t1 join t2 on t1.c1=t2.c1 join t3 on t3.c2=t1.c1 join t4 on t4.c1=t1.c1`, nil},
		{"multi joins too many", "select * from t1 join t2 on t1.c1=t2.c1 join t3 on t3.c2=t1.c1 join t4 on t4.c1=t1.c1 join t5 on t5.c2=t1.c1", ErrMultiJoinNotSupported},
		{"cartesian join 1", "select * from t3, t4", ErrSelectFromMultipleTables},
		// join with condition
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

	ctx := DatabaseContext{
		Tables: map[string]TableContext{
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
			"t5": {
				Columns:     []string{"c1", "c2", "c3"},
				PrimaryKeys: []string{"c1"},
			},
		},
		Actions: map[string]ActionContext{
			"action1": {
				"$this": nil,
				"$that": nil,
			},
			"action2": {
				"$here":  nil,
				"$there": nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := NewErrorHandler(1)
			el := newSqliteErrorListener(eh)
			err := ParseRawSQL(tt.input, 1, "action1", ctx, el, *trace, true)

			if err == nil && tt.wantError == nil {
				return
			}

			if err != nil && tt.wantError != nil {
				// TODO: errors.Is?
				if strings.Contains(err.Error(), tt.wantError.Error()) {
					return
				}
				t.Errorf("ParseRawSQL() expected error: %s, got %s", tt.wantError, err)
				return
			}

			t.Errorf("ParseRawSQL() expected: %s, got %s", tt.wantError, err)
		})
	}
}
