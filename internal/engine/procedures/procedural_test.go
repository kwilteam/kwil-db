package procedures_test

import (
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	procedural "github.com/kwilteam/kwil-db/internal/engine/procedures"
)

// this test is mostly to test two things:
// 1. proper interface conversion among the different visitors
// 2. strong typing for procedures
// We won't actually verify the outputs of the generated code for now.

func Test_Procedures(t *testing.T) {
	type testcase struct {
		name      string
		procedure *types.Procedure
		wantErr   bool
	}

	tests := []testcase{
		// {
		// 	name: "simple",
		// 	procedure: procedure("simple").
		// 		withParams("$id:int").
		// 		withBody(`
		// 		$id2 int := $id + 1;
		// 		`).
		// 		build(),
		// },
		// {
		// 	name: "adding text",
		// 	procedure: procedure("adding_text").
		// 		withParams().
		// 		withBody(`
		// 		$id2 int := 1 + '1';
		// 		`).build(),
		// 	wantErr: true,
		// },
		// {
		// 	name: "returning single value",
		// 	procedure: procedure("returning_single_value").
		// 		withParams().
		// 		withBody(`
		// 		return 1;
		// 		`).withReturns("int").build(),
		// },
		// {
		// 	name: "returning multiple values",
		// 	procedure: procedure("returning_multiple_values").
		// 		withParams().
		// 		withBody(`
		// 		return 1, 'hello';
		// 		`).withReturns("int", "text").build(),
		// },
		// {
		// 	name: "returning table",
		// 	procedure: procedure("returning_table").
		// 		withParams().
		// 		withBody(`
		// 		return SELECT id, name FROM users;
		// 		`).withReturns("table(id int, name text)").build(),
		// },
		// {
		// 	name: "for loop select stmt",
		// 	procedure: procedure("for_loop").
		// 		withParams().
		// 		withBody(`
		// 		$arr int[];
		// 		FOR $row IN SELECT id, name FROM users {
		// 			$arr := array_append($arr, $row.id);
		// 		}

		// 		return $arr;
		// 		`).withReturns("int[]").build(),
		// },
		// {
		// 	name: "for loop select to an invalid type",
		// 	procedure: procedure("for_loop_invalid_type").
		// 		withParams().
		// 		withBody(`
		// 		$arr text[];
		// 		FOR $row IN SELECT id, name FROM users {
		// 			$arr := array_append($arr, $row.id);
		// 		}

		// 		return $arr;
		// 		`).
		// 		withReturns("int").build(),
		// 	wantErr: true,
		// },
		// {
		// 	name: "for loop over array",
		// 	procedure: procedure("for_loop_array").
		// 		withParams("$arr:text[]").
		// 		withBody(`
		// 		FOR $i IN $arr {
		// 			return $i;
		// 		}
		// 		`).
		// 		withReturns("text").
		// 		build(),
		// },
		// {
		// 	name: "for loop over array with invalid type",
		// 	procedure: procedure("for_loop_array_invalid_type").
		// 		withParams("$arr:text[]").
		// 		withBody(`
		// 		FOR $i IN $arr {
		// 			return $i;
		// 		}
		// 		`).
		// 		withReturns("int").
		// 		build(),
		// 	wantErr: true,
		// },
		// {
		// 	name: "for loop over range",
		// 	procedure: procedure("for_loop_range").
		// 		withParams("$start:int", "$end:int").
		// 		withBody(`
		// 		$arr int[];
		// 		FOR $i IN $start:$end {
		// 			$arr := array_append($arr, $i);
		// 		}

		// 		return $arr;
		// 		`).
		// 		withReturns("int[]").
		// 		build(),
		// },
		// {
		// 	name: "for loop over a procedure call",
		// 	procedure: procedure("for_loop_procedure_call").
		// 		withParams().
		// 		withBody(`
		// 		for $row in get_users() {
		// 			if $row.id == 10 {
		// 				break;
		// 			}
		// 		}
		// 		`).
		// 		build(),
		// },
		// {
		// 	name: "array manipulations",
		// 	procedure: procedure("array_manipulations").
		// 		withParams("$name:text").
		// 		withBody(`
		// 		$name_arr text[] := ['name1', 'name2'];
		// 		$name_arr := array_append($name_arr, $name);
		// 		$name_arr2 text[] := array_cat($name_arr, ['name3', 'name4']);
		// 		$len int := array_length($name_arr2);

		// 		return $name_arr2[$len - 1];
		// 		`).
		// 		withReturns("text").
		// 		build(),
		// },
		// {
		// 	name: "invalid array manipulation",
		// 	procedure: procedure("invalid_array_manipulations").
		// 		withParams("$name:text").
		// 		withBody(`
		// 		$name_arr text := 'name1';
		// 		$name_arr := array_append($name_arr, $name);
		// 		`).
		// 		withReturns("text").
		// 		build(),
		// 	wantErr: true,
		// },
		// {
		// 	name: "bare procedure call",
		// 	procedure: procedure("bare_procedure_call").
		// 		withParams().
		// 		withBody(`
		// 		create_user('name', 10, 'address');
		// 		`).
		// 		build(),
		// },
		// {
		// 	name: "assigning two variables from a procedure call",
		// 	procedure: procedure("assigning_two_variables").
		// 		withParams().
		// 		withBody(`
		// 		$id int;
		// 		$name text;
		// 		$id, $name := get_user(1);
		// 		return $id, $name;
		// 		`).
		// 		withReturns("int", "text").
		// 		build(),
		// },
		// {
		// 	name: "assigning two variables from a procedure call with invalid type",
		// 	procedure: procedure("assigning_two_variables_invalid_type").
		// 		withParams().
		// 		withBody(`
		// 		$id text;
		// 		$name text;
		// 		$id, $name := get_user(1);
		// 		`).
		// 		build(),
		// 	wantErr: true,
		// },
		// {
		// 	name: "misplaced break",
		// 	procedure: procedure("misplaced_break").
		// 		withParams().
		// 		withBody(`
		// 		break;
		// 		`).
		// 		build(),
		// 	wantErr: true,
		// },
		// {
		// 	name: "return next",
		// 	procedure: procedure("return_next").
		// 		withParams().
		// 		withBody(`
		// 		for $row in SELECT id FROM users {
		// 			return next $row;
		// 		}
		// 		`).
		// 		withReturns("table(id int)").
		// 		build(),
		// },
		// {
		// 	name: "if elseif else",
		// 	procedure: procedure("if_elseif_else").
		// 		withParams("$id:int").
		// 		withBody(`
		// 		if $id > 10 {
		// 			return 1;
		// 		} elseif $id < 10 {
		// 			return 2;
		// 		} else {
		// 			return 3;
		// 		}
		// 		`).
		// 		withReturns("int").
		// 		build(),
		// },
		// {
		// 	name: "multiple ifs with different types",
		// 	procedure: procedure("multiple_ifs").
		// 		withParams("$id:int").
		// 		withBody(`
		// 		if $id > 10 {
		// 			return 1;
		// 		} elseif $id < 10 {
		// 			return 3, 4;
		// 		} else {
		// 			return 3;
		// 		}
		// 		`).
		// 		withReturns("int").
		// 		build(),
		// 	wantErr: true,
		// },
		{
			name: "sql stmt",
			procedure: procedure("sql_stmt").
				withParams().
				withBody(`
				INSERT INTO users (name, age, address) VALUES (@txid, 10, @caller);
				`).
				build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := procedural.GeneratePLPGSQL(&types.Schema{
				Tables:     testTables,
				Procedures: append(testProcedures, tt.procedure),
			}, "test_schema", "ctx", execution.PgSessionVars)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePLPGSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

var (
	testTables = []*types.Table{
		{
			Name: "users",
			Columns: []*types.Column{
				{
					Name: "id",
					Type: types.IntType,
					Attributes: []*types.Attribute{
						{
							Type: types.PRIMARY_KEY,
						},
					},
				},
				{
					Name: "name",
					Type: types.TextType,
				},
				{
					Name: "age",
					Type: types.IntType,
				},
				{
					Name: "address",
					Type: types.TextType,
				},
			},
		},
		{
			Name: "posts",
			Columns: []*types.Column{
				{
					Name: "id",
					Type: types.IntType,
					Attributes: []*types.Attribute{
						{
							Type: types.PRIMARY_KEY,
						},
					},
				},
				{
					Name: "title",
					Type: types.TextType,
				},
				{
					Name: "content",
					Type: types.TextType,
				},
				{
					Name: "author_id",
					Type: types.IntType,
				},
			},
		},
	}

	testProcedures = []*types.Procedure{
		// procedure("get_users").
		// 	withBody(`
		// 	return SELECT id, name FROM users;
		// 	`).
		// 	withReturns("table(id int, name text)").
		// 	build(),
		// procedure("get_user").
		// 	withParams("$id:int").
		// 	withBody(`
		// 	for $row in SELECT id, name FROM users WHERE id = $id {
		// 		return $row.id, $row.name;
		// 	}
		// 	`).
		// 	withReturns("int", "text").
		// 	build(),
		// procedure("create_user").
		// 	withParams("$name:text", "$age:int", "$address:text").
		// 	withBody(`
		// 	INSERT INTO users (name, age, address) VALUES ($name, $age, $address);
		// 	`).
		// 	build(),
	}
)

// procedureBuilder is a utility to build a procedure for testing.
type procedureBuilder struct {
	name    string
	params  []string
	body    string
	returns []string
}

func procedure(name string) *procedureBuilder {
	return &procedureBuilder{
		name: name,
	}
}

func (pb *procedureBuilder) withParams(params ...string) *procedureBuilder {
	pb.params = params
	return pb
}

func (pb *procedureBuilder) withBody(body string) *procedureBuilder {
	pb.body = body
	return pb
}

func (pb *procedureBuilder) withReturns(returns ...string) *procedureBuilder {
	pb.returns = returns
	return pb
}

func (pb *procedureBuilder) build() *types.Procedure {
	params := make([]*types.ProcedureParameter, len(pb.params))
	for i, p := range pb.params {
		strs := strings.Split(p, ":")

		isArray := false
		if strings.HasSuffix(strs[1], "[]") {
			isArray = true
			strs[1] = strs[1][:len(strs[1])-2]
		}

		params[i] = &types.ProcedureParameter{
			Name: strs[0],
			Type: &types.DataType{
				Name:    strs[1],
				IsArray: isArray,
			},
		}
	}

	procReturn := &types.ProcedureReturn{}

	if len(pb.returns) == 1 && strings.HasPrefix(pb.returns[0], "table") {
		// is of type table(name type, name2 type2, ...)
		// must parse the name and type
		strs := strings.Split(pb.returns[0], "(")
		strs = strings.Split(strs[1], ")")

		for _, s := range strings.Split(strs[0], ",") {
			s = strings.TrimSpace(s)
			sstrs := strings.Split(s, " ")

			isArray := false
			if strings.HasSuffix(sstrs[1], "[]") {
				isArray = true
				sstrs[1] = sstrs[1][:len(sstrs[1])-2]
			}

			procReturn.Table = append(procReturn.Table, &types.NamedType{
				Name: sstrs[0],
				Type: &types.DataType{
					Name:    sstrs[1],
					IsArray: isArray,
				},
			})
		}
	} else {
		returns := make([]*types.DataType, len(pb.returns))

		for i, r := range pb.returns {
			isArray := false
			if strings.HasSuffix(r, "[]") {
				isArray = true
				r = r[:len(r)-2]
			}

			returns[i] = &types.DataType{
				Name:    r,
				IsArray: isArray,
			}
		}
		procReturn.Types = returns
	}

	return &types.Procedure{
		Name:       pb.name,
		Public:     true,
		Parameters: params,
		Body:       pb.body,
		Returns:    procReturn,
	}
}
