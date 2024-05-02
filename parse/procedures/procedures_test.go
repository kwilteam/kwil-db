package procedures_test

import (
	"fmt"
	"strings"
	"testing"

	coreTypes "github.com/kwilteam/kwil-db/core/types"
	procedural "github.com/kwilteam/kwil-db/parse/procedures"
	"github.com/kwilteam/kwil-db/parse/types"
	"github.com/stretchr/testify/require"
)

// flag that can be set to false for quick debugging.
// If true, it will deploy 3 helper procedures to the schema
// that can be used in the tests.
var deployHelperProcedures = true

// this test is meant to test the error outputs of certain procedure
// errors. These include syntax errors, types errors, etc.
func Test_Procedures(t *testing.T) {
	type testcase struct {
		name      string
		procedure *coreTypes.Procedure
		wantErr   error
	}

	tests := []testcase{
		{
			name: "simple",
			procedure: procedure("simple").
				withParams("$id:int").
				withBody(`
				$id2 int := $id + 1;
				`).
				build(),
		},
		{
			name: "adding text",
			procedure: procedure("adding_text").
				withParams().
				withBody(`
				$id2 int := 1 + '1';
				`).build(),
			wantErr: types.ErrArithmeticType,
		},
		{
			name: "returning single value",
			procedure: procedure("returning_single_value").
				withParams().
				withBody(`
				return 1;
				`).withReturns("int").build(),
		},
		{
			name: "returning multiple values",
			procedure: procedure("returning_multiple_values").
				withParams().
				withBody(`
				return 1, 'hello';
				`).withReturns("int", "text").build(),
		},
		{
			name: "returning table",
			procedure: procedure("returning_table").
				withParams().
				withBody(`
				return SELECT id, name FROM users;
				`).withReturns("table(id int, name text)").build(),
		},
		{
			name: "for loop select stmt",
			procedure: procedure("for_loop").
				withParams().
				withBody(`
				$arr int[];
				FOR $row IN SELECT id, name FROM users {
					$arr := array_append($arr, $row.id);
				}

				return $arr;
				`).withReturns("int[]").build(),
		},
		{
			name: "for loop select to an invalid type",
			procedure: procedure("for_loop_invalid_type").
				withParams().
				withBody(`
				$arr text[];
				FOR $row IN SELECT id, name FROM users {
					$arr := array_append($arr, $row.id);
				}

				return $arr;
				`).
				withReturns("int").build(),
			wantErr: types.ErrArrayType,
		},
		{
			name: "for loop over array",
			procedure: procedure("for_loop_array").
				withParams("$arr:text[]").
				withBody(`
				FOR $i IN $arr {
					return $i;
				}
				`).
				withReturns("text").
				build(),
		},
		{
			name: "for loop over array return invalid type",
			procedure: procedure("for_loop_array_invalid_type").
				withParams("$arr:text[]").
				withBody(`
				FOR $i IN $arr {
					return $i;
				}
				`).
				withReturns("int").
				build(),
			wantErr: types.ErrAssignment,
		},
		{
			name: "for loop over range",
			procedure: procedure("for_loop_range").
				withParams("$start:int", "$end:int").
				withBody(`
				$arr int[];
				FOR $i IN $start:$end {
					$arr := array_append($arr, $i);
				}

				return $arr;
				`).
				withReturns("int[]").
				build(),
		},
		{
			name: "for loop over a procedure call",
			procedure: procedure("for_loop_procedure_call").
				withParams().
				withBody(`
				for $row in get_users() {
					if $row.id == 10 {
						break;
					}
				}
				`).
				build(),
		},
		{
			name: "array manipulations",
			procedure: procedure("array_manipulations").
				withParams("$name:text").
				withBody(`
				$name_arr text[] := ['name1', 'name2'];
				$name_arr := array_append($name_arr, $name);
				$name_arr2 text[] := array_cat($name_arr, ['name3', 'name4']);
				$len int := array_length($name_arr2);

				return $name_arr2[$len - 1];
				`).
				withReturns("text").
				build(),
		},
		{
			name: "invalid array manipulation",
			procedure: procedure("invalid_array_manipulations").
				withParams("$name:text").
				withBody(`
				$name_arr text := 'name1';
				$name_arr2 text[] := array_append($name_arr, $name);
				`).
				build(),
			wantErr: types.ErrParamType,
		},
		{
			name: "bare procedure call",
			procedure: procedure("bare_procedure_call").
				withParams().
				withBody(`
				create_user('name', 10, 'address');
				`).
				build(),
		},
		{
			name: "assigning two variables from a procedure call",
			procedure: procedure("assigning_two_variables").
				withParams().
				withBody(`
				$id int;
				$name text;
				$id, $name := get_user(1);
				return $id, $name;
				`).
				withReturns("int", "text").
				build(),
		},
		{
			name: "assigning two variables from a procedure call with invalid type",
			procedure: procedure("assigning_two_variables_invalid_type").
				withParams().
				withBody(`
				$id text;
				$name text;
				$id, $name := get_user(1);
				`).
				build(),
			wantErr: types.ErrAssignment,
		},
		{
			name: "misplaced break",
			procedure: procedure("misplaced_break").
				withParams().
				withBody(`
				break;
				`).
				build(),
			wantErr: types.ErrBreakUsedOutsideOfLoop,
		},
		{
			name: "return next",
			procedure: procedure("return_next").
				withParams().
				withBody(`
				for $row in SELECT id FROM users {
					return next $row.id;
				}
				`).
				withReturns("table(id int)").
				build(),
		},
		{
			name: "if elseif else",
			procedure: procedure("if_elseif_else").
				withParams("$id:int").
				withBody(`
				if $id > 10 {
					return 1;
				} elseif $id < 10 {
					return 2;
				} else {
					return 3;
				}
				`).
				withReturns("int").
				build(),
		},
		{
			name: "multiple ifs with different types",
			procedure: procedure("multiple_ifs").
				withParams("$id:int").
				withBody(`
				if $id > 10 {
					return 1;
				} elseif $id < 10 {
					return 3, 4;
				} else {
					return 3;
				}
				`).
				withReturns("int").
				build(),
			wantErr: types.ErrReturnCount,
		},
		{
			name: "sql stmt",
			procedure: procedure("sql_stmt").
				withParams().
				withBody(`
				INSERT INTO users (name, age, address) VALUES (@txid, 10, @caller);
				`).
				build(),
		},
		{
			name: "return single user",
			procedure: procedure("return_single_user").
				withParams("$id:int").
				withBody(`
				for $row in SELECT id, name FROM users WHERE id = $id {
					return $row.id, $row.name;
				}
				error('user not found');
				`).withReturns("int", "text").build(),
		},
		{
			name: "procedure call returning table",
			procedure: procedure("proc_call").
				withParams().
				withBody(`
				$res int := get_users();
				`).
				build(),
			wantErr: types.ErrAssignment,
		},
		{
			name: "procedure call with anonymous receiver",
			procedure: procedure("proc_call_anonymous_receiver").
				withParams().
				withBody(`
				$id int;
				$id, _ := get_user(1);
				`).
				build(),
		},
		{
			name: "null comparison",
			procedure: procedure("null_comparison").
				withParams("$name:text").
				withBody(`
				if $name == null {
				}
				`).build(),
		},
		{
			name: "sql- invalid join syntax",
			procedure: procedure("invalid_join").
				withParams().
				withBody(`
				SELECT * FROM users INNER J;
				`).build(),
			wantErr: types.ErrSyntaxError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			procs := []*coreTypes.Procedure{tt.procedure}
			if deployHelperProcedures {
				procs = append(procs, testProcedures...)
			}

			_, parseErrs, err := procedural.AnalyzeProcedures(&coreTypes.Schema{
				Tables:     testTables,
				Procedures: procs,
			}, "test_schema", nil)
			require.NoError(t, err) // ensure no error is returned, and that they were caught by the error listener
			// we want to check parse errors in this test too
			if tt.wantErr != nil {
				require.NotNil(t, parseErrs)
				// we want to ensure that the first error is the one we expect
				firstErr := parseErrs[0]
				require.Contains(t, firstErr.Error(), tt.wantErr.Error())
			} else {
				require.Empty(t, parseErrs)
			}
		})
	}
}

var (
	testTables = []*coreTypes.Table{
		{
			Name: "users",
			Columns: []*coreTypes.Column{
				{
					Name: "id",
					Type: coreTypes.IntType,
					Attributes: []*coreTypes.Attribute{
						{
							Type: coreTypes.PRIMARY_KEY,
						},
					},
				},
				{
					Name: "name",
					Type: coreTypes.TextType,
				},
				{
					Name: "age",
					Type: coreTypes.IntType,
				},
				{
					Name: "address",
					Type: coreTypes.TextType,
				},
			},
		},
		{
			Name: "posts",
			Columns: []*coreTypes.Column{
				{
					Name: "id",
					Type: coreTypes.IntType,
					Attributes: []*coreTypes.Attribute{
						{
							Type: coreTypes.PRIMARY_KEY,
						},
					},
				},
				{
					Name: "title",
					Type: coreTypes.TextType,
				},
				{
					Name: "content",
					Type: coreTypes.TextType,
				},
				{
					Name: "author_id",
					Type: coreTypes.IntType,
				},
			},
		},
	}

	testProcedures = []*coreTypes.Procedure{
		procedure("get_users").
			withBody(`
			return SELECT id, name FROM users;
			`).
			withReturns("table(id int, name text)").
			build(),
		procedure("get_user").
			withParams("$id:int").
			withBody(`
			for $row in SELECT id, name FROM users WHERE id = $id {
				return $row.id, $row.name;
			}
			`).
			withReturns("int", "text").
			build(),
		procedure("create_user").
			withParams("$name:text", "$age:int", "$address:text").
			withBody(`
			INSERT INTO users (name, age, address) VALUES ($name, $age, $address);
			`).
			build(),
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

func (pb *procedureBuilder) build() *coreTypes.Procedure {
	params := make([]*coreTypes.ProcedureParameter, len(pb.params))
	for i, p := range pb.params {
		strs := strings.Split(p, ":")

		isArray := false
		if strings.HasSuffix(strs[1], "[]") {
			isArray = true
			strs[1] = strs[1][:len(strs[1])-2]
		}

		params[i] = &coreTypes.ProcedureParameter{
			Name: strs[0],
			Type: &coreTypes.DataType{
				Name:    strs[1],
				IsArray: isArray,
			},
		}
	}

	procReturn := &coreTypes.ProcedureReturn{}

	if len(pb.returns) == 1 && strings.HasPrefix(pb.returns[0], "table") {
		procReturn.IsTable = true
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

			procReturn.Fields = append(procReturn.Fields, &coreTypes.NamedType{
				Name: sstrs[0],
				Type: &coreTypes.DataType{
					Name:    sstrs[1],
					IsArray: isArray,
				},
			})
		}
	} else {
		returns := make([]*coreTypes.NamedType, len(pb.returns))

		for i, r := range pb.returns {
			isArray := false
			if strings.HasSuffix(r, "[]") {
				isArray = true
				r = r[:len(r)-2]
			}

			returns[i] = &coreTypes.NamedType{
				Name: fmt.Sprintf("ret%d", i),
				Type: &coreTypes.DataType{
					Name:    r,
					IsArray: isArray,
				},
			}
		}
		procReturn.Fields = returns
	}

	return &coreTypes.Procedure{
		Name:       pb.name,
		Public:     true,
		Parameters: params,
		Body:       pb.body,
		Returns:    procReturn,
	}
}
