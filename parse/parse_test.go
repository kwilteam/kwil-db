package parse_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/order"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/stretchr/testify/require"
)

// Test_Kuneiform tests the Kuneiform parser.
func Test_Kuneiform(t *testing.T) {
	type testCase struct {
		name string
		kf   string
		want *types.Schema
	}

	tests := []testCase{
		{
			name: "simple schema",
			kf: `
			database mydb;

			table users {
				id int primary key not null,
				username text not null unique minlen(5) maxlen(32)
			}

			action create_user ($id, $username) public {
				insert into users (id, username) values ($id, $username);
			}

			procedure get_username ($id int) public view RETURNS (name text) {
				return select username from users where id = $id; // this is a comment
			}
			`,
			want: &types.Schema{
				Name: "mydb",
				Tables: []*types.Table{
					tblUsers,
				},
				Actions: []*types.Action{
					{
						Name: "create_user",
						Parameters: []string{
							"$id",
							"$username",
						},
						Public: true,
						Body:   `insert into users (id, username) values ($id, $username);`,
					},
				},
				Procedures: []*types.Procedure{
					{
						Name: "get_username",
						Parameters: []*types.ProcedureParameter{
							{
								Name: "$id",
								Type: types.IntType,
							},
						},
						Public: true,
						Modifiers: []types.Modifier{
							types.ModifierView,
						},
						Body: `return select username from users where id = $id;`,
						Returns: &types.ProcedureReturn{Fields: []*types.NamedType{
							{
								Name: "name",
								Type: types.TextType,
							},
						}},
					},
				},
			},
		},
		{
			name: "foreign key and index",
			kf: `
			database mydb;

			table users {
				id int primary key not null,
				username text not null unique minlen(5) maxlen(32)
			}

			table posts {
				id int primary key,
				author_id int not null,
				foreign key (author_id) references users (id) on delete cascade on update cascade,
				#idx index(author_id)
			}
			`,
			want: &types.Schema{
				Name: "mydb",
				Tables: []*types.Table{
					tblUsers,
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
								Name: "author_id",
								Type: types.IntType,
								Attributes: []*types.Attribute{
									{
										Type: types.NOT_NULL,
									},
								},
							},
						},
						Indexes: []*types.Index{
							{
								Name:    "idx",
								Type:    types.BTREE,
								Columns: []string{"author_id"},
							},
						},
						ForeignKeys: []*types.ForeignKey{
							{
								ChildKeys:   []string{"author_id"},
								ParentTable: "users",
								ParentKeys:  []string{"id"},
								Actions: []*types.ForeignKeyAction{
									{
										On: types.ON_DELETE,
										Do: types.DO_CASCADE,
									},
									{
										On: types.ON_UPDATE,
										Do: types.DO_CASCADE,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "procedure returns table",
			kf: `
			database mydb;

			procedure get_users() public view RETURNS table(id int) {
				return select id from users;
			}
			`,
			want: &types.Schema{
				Name: "mydb",
				Procedures: []*types.Procedure{
					{
						Name:   "get_users",
						Public: true,
						Modifiers: []types.Modifier{
							types.ModifierView,
						},
						Body: `return select id from users;`,
						Returns: &types.ProcedureReturn{
							IsTable: true,
							Fields: []*types.NamedType{
								{
									Name: "id",
									Type: types.IntType,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "use",
			kf: `
			database mydb;

			uSe myext AS db1;
			use myext {
				a: 'b',
				c: 1
			} aS db2;
			`,
			want: &types.Schema{
				Name: "mydb",
				Extensions: []*types.Extension{
					{
						Name:  "myext",
						Alias: "db1",
					},
					{
						Name: "myext",
						Initialization: []*types.ExtensionConfig{
							{
								Key:   "a",
								Value: "'b'",
							},
							{
								Key:   "c",
								Value: "1",
							},
						},
						Alias: "db2",
					},
				},
			},
		},
		{
			name: "annotations",
			kf: `
			database mydb;

			@kgw(authn='true')
			procedure get_users() public view {}
			`,
			want: &types.Schema{
				Name: "mydb",
				Procedures: []*types.Procedure{
					{
						Name:   "get_users",
						Public: true,
						Modifiers: []types.Modifier{
							types.ModifierView,
						},
						Annotations: []string{"@kgw(authn='true')"},
					},
				},
			},
		},
		{
			name: "all possible constraints",
			kf: `
			database mydb;

			table other_users {
				id int primary key,
				username text not null unique minlen(5) maxlen(32),
				age int max(100) min(18) default(18)
			}

			table users {
				id int primary key,
				username text not null unique minlen(5) maxlen(32),
				age int max(100) min(18) default(18),
				bts blob default(0x00),
				foreign key (id) references other_users (id) on delete cascade on update set null,
				foreign key (username) references other_users (username) on delete set default on update no action,
				foreign key (age) references other_users (age) on delete restrict
			}

			table other_uses {
				id int primary key,
				username text unique,
				age int unique
			}
			`,
			want: &types.Schema{
				Name: "mydb",
				Tables: []*types.Table{
					{
						Name: "other_users",
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
								Name: "username",
								Type: types.TextType,
								Attributes: []*types.Attribute{
									{
										Type: types.NOT_NULL,
									},
									{
										Type: types.UNIQUE,
									},
									{
										Type:  types.MIN_LENGTH,
										Value: "5",
									},
									{
										Type:  types.MAX_LENGTH,
										Value: "32",
									},
								},
							},
							{
								Name: "age",
								Type: types.IntType,
								Attributes: []*types.Attribute{
									{
										Type:  types.MAX,
										Value: "100",
									},
									{
										Type:  types.MIN,
										Value: "18",
									},
									{
										Type:  types.DEFAULT,
										Value: "18",
									},
								},
							},
						},
					},
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
								Name: "username",
								Type: types.TextType,
								Attributes: []*types.Attribute{
									{
										Type: types.NOT_NULL,
									},
									{
										Type: types.UNIQUE,
									},
									{
										Type:  types.MIN_LENGTH,
										Value: "5",
									},
									{
										Type:  types.MAX_LENGTH,
										Value: "32",
									},
								},
							},
							{
								Name: "age",
								Type: types.IntType,
								Attributes: []*types.Attribute{
									{
										Type:  types.MAX,
										Value: "100",
									},
									{
										Type:  types.MIN,
										Value: "18",
									},
									{
										Type:  types.DEFAULT,
										Value: "18",
									},
								},
							},
							{
								Name: "bts",
								Type: types.BlobType,
								Attributes: []*types.Attribute{
									{
										Type:  types.DEFAULT,
										Value: "0x00",
									},
								},
							},
						},
						ForeignKeys: []*types.ForeignKey{
							{
								ChildKeys:   []string{"id"},
								ParentTable: "other_users",
								ParentKeys:  []string{"id"},
								Actions: []*types.ForeignKeyAction{
									{
										On: types.ON_DELETE,
										Do: types.DO_CASCADE,
									},
									{
										On: types.ON_UPDATE,
										Do: types.DO_SET_NULL,
									},
								},
							},
							{
								ChildKeys:   []string{"username"},
								ParentTable: "other_users",
								ParentKeys:  []string{"username"},
								Actions: []*types.ForeignKeyAction{
									{
										On: types.ON_DELETE,
										Do: types.DO_SET_DEFAULT,
									},
									{
										On: types.ON_UPDATE,
										Do: types.DO_NO_ACTION,
									},
								},
							},
							{
								ChildKeys:   []string{"age"},
								ParentTable: "other_users",
								ParentKeys:  []string{"age"},
								Actions: []*types.ForeignKeyAction{
									{
										On: types.ON_DELETE,
										Do: types.DO_RESTRICT,
									},
								},
							},
						},
					},
					{
						Name: "other_uses",
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
								Name: "username",
								Type: types.TextType,
								Attributes: []*types.Attribute{
									{
										Type: types.UNIQUE,
									},
								},
							},
							{
								Name: "age",
								Type: types.IntType,
								Attributes: []*types.Attribute{
									{
										Type: types.UNIQUE,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "foreign, no parameters, returns nothing",
			kf: `
			database mydb;

			foreign procedure get_users()
			`,
			want: &types.Schema{
				Name: "mydb",
				ForeignProcedures: []*types.ForeignProcedure{
					{
						Name: "get_users",
					},
				},
			},
		},
		{
			name: "foreign, with parameters, returns unnamed types",
			kf: `
			database mydb;

			foreign procedure get_users(int, text) RETURNS (int, text)
			`,
			want: &types.Schema{
				Name: "mydb",
				ForeignProcedures: []*types.ForeignProcedure{
					{
						Name: "get_users",
						Parameters: []*types.DataType{
							types.IntType,
							types.TextType,
						},
						Returns: &types.ProcedureReturn{
							Fields: []*types.NamedType{
								{
									Name: "col0",
									Type: types.IntType,
								},
								{
									Name: "col1",
									Type: types.TextType,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "foreign, with parameters, returns named types",
			kf: `
			database mydb;

			foreign procedure get_users() RETURNS (id int, name text)
			`,
			want: &types.Schema{
				Name: "mydb",
				ForeignProcedures: []*types.ForeignProcedure{
					{
						Name: "get_users",
						Returns: &types.ProcedureReturn{
							Fields: []*types.NamedType{
								{
									Name: "id",
									Type: types.IntType,
								},
								{
									Name: "name",
									Type: types.TextType,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "foreign,  returns table",
			kf: `
			database mydb;

			foreign   procedure   get_users() RETURNS table(id int)
			`,
			want: &types.Schema{
				Name: "mydb",
				ForeignProcedures: []*types.ForeignProcedure{
					{
						Name: "get_users",
						Returns: &types.ProcedureReturn{
							IsTable: true,
							Fields: []*types.NamedType{
								{
									Name: "id",
									Type: types.IntType,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "named foreign parameters",
			kf: `
			database mydb;

			foreign procedure get_users($id int, $name text) returns (id int, name text)
			`,
			want: &types.Schema{
				Name: "mydb",
				ForeignProcedures: []*types.ForeignProcedure{
					{
						Name: "get_users",
						Parameters: []*types.DataType{
							types.IntType,
							types.TextType,
						},
						Returns: &types.ProcedureReturn{
							Fields: []*types.NamedType{
								{
									Name: "id",
									Type: types.IntType,
								},
								{
									Name: "name",
									Type: types.TextType,
								},
							},
						},
					},
				},
			},
		},
		{
			// this test tries to break case sensitivity in every way possible
			name: "case insensitive",
			kf: `
			database myDB;
			
			table UsErS {
				iD inT pRimaRy kEy nOt nUll
			}

			table posts {
				id int primary key,
				author_id int not null,
				ForEign key (author_ID) references usErs (Id) On delEte cAscade on Update cascadE,
				#iDx inDex(author_iD)
			}
			
			uSe myeXt As dB1;

			pRoceDure get_Users($nAme tExt) Public viEw ReTURNS tablE(iD iNt) {
				return select id from users; // this wont actually get parsed in this test
			}

			fOreign proceduRe get_othEr_Users($Id inT, $nAme Text) RETURNS table(iD inT, Name tExt)

			@kGw( autHn='tRue' )
			AcTion create_User ($Id, $usErname) Public {
				insert into users (id, username) values ($id, $username);
			}
			`,
			want: &types.Schema{
				Name: "mydb",
				Tables: []*types.Table{
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
									{
										Type: types.NOT_NULL,
									},
								},
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
								Name: "author_id",
								Type: types.IntType,
								Attributes: []*types.Attribute{
									{
										Type: types.NOT_NULL,
									},
								},
							},
						},
						Indexes: []*types.Index{
							{
								Name:    "idx",
								Type:    types.BTREE,
								Columns: []string{"author_id"},
							},
						},
						ForeignKeys: []*types.ForeignKey{
							{
								ChildKeys:   []string{"author_id"},
								ParentTable: "users",
								ParentKeys:  []string{"id"},
								Actions: []*types.ForeignKeyAction{
									{
										On: types.ON_DELETE,
										Do: types.DO_CASCADE,
									},
									{
										On: types.ON_UPDATE,
										Do: types.DO_CASCADE,
									},
								},
							},
						},
					},
				},
				Extensions: []*types.Extension{
					{
						Name:  "myext",
						Alias: "db1",
					},
				},
				Procedures: []*types.Procedure{
					{
						Name:   "get_users",
						Public: true,
						Modifiers: []types.Modifier{
							types.ModifierView,
						},
						Parameters: []*types.ProcedureParameter{
							{
								Name: "$name",
								Type: types.TextType,
							},
						},
						Returns: &types.ProcedureReturn{
							IsTable: true,
							Fields: []*types.NamedType{
								{
									Name: "id",
									Type: types.IntType,
								},
							},
						},
						Body: `return select id from users;`, // comments will not be parsed
					},
				},
				ForeignProcedures: []*types.ForeignProcedure{
					{
						Name: "get_other_users",
						Parameters: []*types.DataType{
							types.IntType,
							types.TextType,
						},
						Returns: &types.ProcedureReturn{
							IsTable: true,
							Fields: []*types.NamedType{
								{
									Name: "id",
									Type: types.IntType,
								},
								{
									Name: "name",
									Type: types.TextType,
								},
							},
						},
					},
				},
				Actions: []*types.Action{
					{
						Annotations: []string{"@kgw(authn='tRue')"},
						Name:        "create_user",
						Parameters: []string{
							"$id",
							"$username",
						},
						Public: true,
						Body:   `insert into users (id, username) values ($id, $username);`,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parse.ParseSchema([]byte(tt.kf))
			require.NoError(t, err)
			require.NoError(t, res.ParseErrs.Err())

			require.EqualValues(t, tt.want, res.Schema)

			// we will also test that the schemas were properly cleaned.
			// we test this by copying the schema to a new schema, cleaning the new schema, and comparing the two.
			bts, err := json.Marshal(res.Schema)
			require.NoError(t, err)

			var got2 types.Schema
			err = json.Unmarshal(bts, &got2)
			require.NoError(t, err)

			err = got2.Clean()
			require.NoError(t, err)

			got2.Owner = nil // unmarshal sets Owner to empty array, so we need to set it to nil to compare

			require.EqualValues(t, res.Schema, &got2)
		})
	}
}

// some default tables and procedures for testing
var (
	tblUsers = &types.Table{
		Name: "users",
		Columns: []*types.Column{
			{
				Name: "id",
				Type: types.IntType,
				Attributes: []*types.Attribute{
					{
						Type: types.PRIMARY_KEY,
					},
					{
						Type: types.NOT_NULL,
					},
				},
			},
			{
				Name: "username",
				Type: types.TextType,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
					{
						Type: types.UNIQUE,
					},
					{
						Type:  types.MIN_LENGTH,
						Value: "5",
					},
					{
						Type:  types.MAX_LENGTH,
						Value: "32",
					},
				},
			},
		},
	}

	tblPosts = &types.Table{
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
				Name: "author_id",
				Type: types.IntType,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
				},
			},
		},
		Indexes: []*types.Index{
			{
				Name:    "idx",
				Type:    types.BTREE,
				Columns: []string{"author_id"},
			},
		},
		ForeignKeys: []*types.ForeignKey{
			{
				ChildKeys:   []string{"author_id"},
				ParentTable: "users",
				ParentKeys:  []string{"id"},
				Actions: []*types.ForeignKeyAction{
					{
						On: types.ON_DELETE,
						Do: types.DO_CASCADE,
					},
					{
						On: types.ON_UPDATE,
						Do: types.DO_CASCADE,
					},
				},
			},
		},
	}

	procGetAllUserIds = &types.Procedure{
		Name:   "get_all_user_ids",
		Public: true,
		Modifiers: []types.Modifier{
			types.ModifierView,
		},
		Returns: &types.ProcedureReturn{
			IsTable: true,
			Fields: []*types.NamedType{
				{
					Name: "id",
					Type: types.IntType,
				},
			},
		},
		Body: `return select id from users;`,
	}
)

func Test_Procedure(t *testing.T) {
	type testCase struct {
		name string
		proc string
		// inputs should be a map of $var to type
		inputs map[string]*types.DataType
		// returns is the expected return type
		// it can be left nil if there is no return type.
		returns *types.ProcedureReturn
		// want is the desired output.
		// Errs should be left nil for this test,
		// and passed in the test case.
		// inputs will automatically be added
		// to the expected output as variables.
		want *parse.ProcedureParseResult
		err  error
	}

	tests := []testCase{
		{
			name: "simple procedure",
			proc: `$a int := 1;`,
			want: &parse.ProcedureParseResult{
				Variables: map[string]*types.DataType{
					"$a": types.IntType,
				},
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtAssign{
						Variable: exprVar("$a"),
						Type:     types.IntType,
						Value:    exprLit(1),
					},
				},
			},
		},
		{
			name: "for loop",
			proc: `
			$found := false;
			for $row in SELECT * FROM users {
				$found := true;
				INSERT INTO posts (id, author_id) VALUES ($row.id, $row.username::int);
			}
			if !$found {
				error('no users found');
			}
			`,
			want: &parse.ProcedureParseResult{
				CompoundVariables: map[string]struct{}{
					"$row": {},
				},
				Variables: map[string]*types.DataType{
					"$found": types.BoolType,
				},
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtAssign{
						Variable: exprVar("$found"),
						Value:    exprLit(false),
					},
					&parse.ProcedureStmtForLoop{
						Receiver: exprVar("$row"),
						LoopTerm: &parse.LoopTermSQL{
							Statement: &parse.SQLStatement{
								SQL: &parse.SelectStatement{
									SelectCores: []*parse.SelectCore{
										{
											Columns: []parse.ResultColumn{
												&parse.ResultColumnWildcard{},
											},
											From: &parse.RelationTable{
												Table: "users",
											},
										},
									},
									// apply default ordering
									Ordering: []*parse.OrderingTerm{
										{
											Expression: &parse.ExpressionColumn{
												Table:  "users",
												Column: "id",
											},
										},
									},
								},
							},
						},
						Body: []parse.ProcedureStmt{
							&parse.ProcedureStmtAssign{
								Variable: exprVar("$found"),
								Value:    exprLit(true),
							},
							&parse.ProcedureStmtSQL{
								SQL: &parse.SQLStatement{
									SQL: &parse.InsertStatement{
										Table:   "posts",
										Columns: []string{"id", "author_id"},
										Values: [][]parse.Expression{
											{
												&parse.ExpressionFieldAccess{
													Record: exprVar("$row"),
													Field:  "id",
												},
												&parse.ExpressionFieldAccess{
													Record: exprVar("$row"),
													Field:  "username",
													Typecastable: parse.Typecastable{
														TypeCast: types.IntType,
													},
												},
											},
										},
									},
								},
							},
						},
					},
					&parse.ProcedureStmtIf{
						IfThens: []*parse.IfThen{
							{
								If: &parse.ExpressionUnary{
									Operator:   parse.UnaryOperatorNot,
									Expression: exprVar("$found"),
								},
								Then: []parse.ProcedureStmt{
									&parse.ProcedureStmtCall{
										Call: &parse.ExpressionFunctionCall{
											Name: "error",
											Args: []parse.Expression{
												exprLit("no users found"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "arrays",
			proc: `
			$arr2 := array_append($arr, 2);
			$arr3 int[] := array_prepend(3, $arr2);
			$arr4 := [4,5];

			$arr5 := array_cat($arr3, $arr4);
			`,
			inputs: map[string]*types.DataType{
				"$arr": types.ArrayType(types.IntType),
			},
			want: &parse.ProcedureParseResult{
				Variables: map[string]*types.DataType{
					"$arr2": types.ArrayType(types.IntType),
					"$arr3": types.ArrayType(types.IntType),
					"$arr4": types.ArrayType(types.IntType),
					"$arr5": types.ArrayType(types.IntType),
				},
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtCall{
						Receivers: []*parse.ExpressionVariable{
							exprVar("$arr2"),
						},
						Call: &parse.ExpressionFunctionCall{
							Name: "array_append",
							Args: []parse.Expression{
								exprVar("$arr"),
								exprLit(2),
							},
						},
					},
					&parse.ProcedureStmtAssign{
						Variable: exprVar("$arr3"),
						Type:     types.ArrayType(types.IntType),
						Value: &parse.ExpressionFunctionCall{
							Name: "array_prepend",
							Args: []parse.Expression{
								exprLit(3),
								exprVar("$arr2"),
							},
						},
					},
					&parse.ProcedureStmtAssign{
						Variable: exprVar("$arr4"),
						Value: &parse.ExpressionMakeArray{
							Values: []parse.Expression{
								exprLit(4),
								exprLit(5),
							},
						},
					},
					&parse.ProcedureStmtCall{
						Receivers: []*parse.ExpressionVariable{
							exprVar("$arr5"),
						},
						Call: &parse.ExpressionFunctionCall{
							Name: "array_cat",
							Args: []parse.Expression{
								exprVar("$arr3"),
								exprVar("$arr4"),
							},
						},
					},
				},
			},
		},
		{
			name: "loop",
			proc: `
			$arr := [1,2,3];
			$rec int;
			for $i in $arr {
				$rec := $i;
			}
			`,
			want: &parse.ProcedureParseResult{
				Variables: map[string]*types.DataType{
					"$arr": types.ArrayType(types.IntType),
					"$rec": types.IntType,
					"$i":   types.IntType,
				},
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtAssign{
						Variable: exprVar("$arr"),
						Value: &parse.ExpressionMakeArray{
							Values: []parse.Expression{
								exprLit(1),
								exprLit(2),
								exprLit(3),
							},
						},
					},
					&parse.ProcedureStmtDeclaration{
						Variable: exprVar("$rec"),
						Type:     types.IntType,
					},
					&parse.ProcedureStmtForLoop{
						Receiver: exprVar("$i"),
						LoopTerm: &parse.LoopTermVariable{
							Variable: exprVar("$arr"),
						},
						Body: []parse.ProcedureStmt{
							&parse.ProcedureStmtAssign{
								Variable: exprVar("$rec"),
								Value:    exprVar("$i"),
							},
						},
					},
				},
			},
		},
		{
			name: "and/or",
			proc: `
			if $a and $b or $c {}
			`,
			inputs: map[string]*types.DataType{
				"$a": types.BoolType,
				"$b": types.BoolType,
				"$c": types.BoolType,
			},
			want: &parse.ProcedureParseResult{
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtIf{
						IfThens: []*parse.IfThen{
							{
								If: &parse.ExpressionLogical{
									Left: &parse.ExpressionLogical{
										Left:     exprVar("$a"),
										Operator: parse.LogicalOperatorAnd,
										Right:    exprVar("$b"),
									},
									Operator: parse.LogicalOperatorOr,
									Right:    exprVar("$c"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "is distinct",
			proc: `
			$a := 1 is distinct from null;
			`,
			want: &parse.ProcedureParseResult{
				Variables: map[string]*types.DataType{
					"$a": types.BoolType,
				},
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtAssign{
						Variable: exprVar("$a"),
						Value: &parse.ExpressionIs{
							Left: exprLit(1),
							Right: &parse.ExpressionLiteral{
								Type: types.NullType,
							},
							Distinct: true,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params []*types.ProcedureParameter
			for _, in := range order.OrderMap(tt.inputs) {
				params = append(params, &types.ProcedureParameter{
					Name: in.Key,
					Type: in.Value,
				})
			}

			proc := &types.Procedure{
				Name:       "test",
				Parameters: params,
				Public:     true,
				Returns:    tt.returns,
				Body:       tt.proc,
			}

			res, err := parse.ParseProcedure(proc, &types.Schema{
				Name: "mydb",
				Tables: []*types.Table{
					tblUsers,
					tblPosts,
				},
				Procedures: []*types.Procedure{
					proc,
					procGetAllUserIds,
				},
			})
			require.NoError(t, err)

			if tt.err != nil {
				require.ErrorIs(t, res.ParseErrs.Err(), tt.err)
				return
			}
			require.NoError(t, res.ParseErrs.Err())

			// set res errs to nil to match test
			res.ParseErrs = nil

			if tt.want.CompoundVariables == nil {
				tt.want.CompoundVariables = make(map[string]struct{})
			}
			if tt.want.Variables == nil {
				tt.want.Variables = make(map[string]*types.DataType)
			}

			// add the inputs to the expected output
			for k, v := range tt.inputs {
				tt.want.Variables[k] = v
			}

			if !deepCompare(tt.want, res) {
				t.Errorf("unexpected output: %s", diff(tt.want, res))
			}
		})
	}
}

// exprVar makes an ExpressionVariable.
func exprVar(n string) *parse.ExpressionVariable {
	if n[0] != '$' && n[0] != '@' {
		panic("TEST ERROR: variable name must start with $ or @")
	}
	pref := parse.VariablePrefix(n[0])

	return &parse.ExpressionVariable{
		Name:   n[1:],
		Prefix: pref,
	}
}

// exprLit makes an ExpressionLiteral.
// it can only make strings and ints
func exprLit(v any) *parse.ExpressionLiteral {
	switch t := v.(type) {
	case int:
		return &parse.ExpressionLiteral{
			Type:  types.IntType,
			Value: int64(t),
		}
	case int64:
		return &parse.ExpressionLiteral{
			Type:  types.IntType,
			Value: t,
		}
	case string:
		return &parse.ExpressionLiteral{
			Type:  types.TextType,
			Value: t,
		}
	case bool:
		return &parse.ExpressionLiteral{
			Type:  types.BoolType,
			Value: t,
		}
	default:
		panic("TEST ERROR: invalid type for literal")
	}
}

func Test_SQL(t *testing.T) {
	type testCase struct {
		name string
		sql  string
		want *parse.SQLStatement
		err  error
	}

	tests := []testCase{
		{
			name: "simple select",
			sql:  "select *, id i, length(username) as name_len from users u where u.id = 1;",
			want: &parse.SQLStatement{
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnWildcard{},
								&parse.ResultColumnExpression{
									Expression: &parse.ExpressionColumn{
										Column: "id",
									},
									Alias: "i",
								},
								&parse.ResultColumnExpression{
									Expression: &parse.ExpressionFunctionCall{
										Name: "length",
										Args: []parse.Expression{
											&parse.ExpressionColumn{
												Column: "username",
											},
										},
									},
									Alias: "name_len",
								},
							},
							From: &parse.RelationTable{
								Table: "users",
								Alias: "u",
							},
							Where: &parse.ExpressionComparison{
								Left: &parse.ExpressionColumn{
									Table:  "u",
									Column: "id",
								},
								Operator: parse.ComparisonOperatorEqual,
								Right: &parse.ExpressionLiteral{
									Type:  types.IntType,
									Value: int64(1),
								},
							},
						},
					},
					// apply default ordering
					Ordering: []*parse.OrderingTerm{
						{
							Expression: &parse.ExpressionColumn{
								Table:  "u",
								Column: "id",
							},
						},
					},
				},
			},
		},
		{
			name: "insert",
			sql: `insert into posts (id, author_id) values (1, 1),
			(2, (SELECT id from users where username = 'user2' LIMIT 1));`,
			want: &parse.SQLStatement{
				SQL: &parse.InsertStatement{
					Table:   "posts",
					Columns: []string{"id", "author_id"},
					Values: [][]parse.Expression{
						{
							&parse.ExpressionLiteral{
								Type:  types.IntType,
								Value: int64(1),
							},
							&parse.ExpressionLiteral{
								Type:  types.IntType,
								Value: int64(1),
							},
						},
						{
							&parse.ExpressionLiteral{
								Type:  types.IntType,
								Value: int64(2),
							},
							&parse.ExpressionSubquery{
								Subquery: &parse.SelectStatement{
									SelectCores: []*parse.SelectCore{
										{
											Columns: []parse.ResultColumn{
												&parse.ResultColumnExpression{
													Expression: &parse.ExpressionColumn{
														Column: "id",
													},
												},
											},
											From: &parse.RelationTable{
												Table: "users",
											},
											Where: &parse.ExpressionComparison{
												Left: &parse.ExpressionColumn{
													Column: "username",
												},
												Operator: parse.ComparisonOperatorEqual,
												Right: &parse.ExpressionLiteral{
													Type:  types.TextType,
													Value: "user2",
												},
											},
										},
									},
									Limit: &parse.ExpressionLiteral{
										Type:  types.IntType,
										Value: int64(1),
									},
									// apply default ordering
									Ordering: []*parse.OrderingTerm{
										{
											Expression: &parse.ExpressionColumn{
												Table:  "users",
												Column: "id",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "select join",
			sql: `SELECT p.id as id, u.username as author FROM posts AS p
			INNER JOIN users AS u ON p.author_id = u.id
			WHERE u.username = 'satoshi' order by u.username DESC NULLS LAST;`,
			want: &parse.SQLStatement{
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnExpression{
									Expression: &parse.ExpressionColumn{
										Column: "id",
										Table:  "p",
									},
									Alias: "id",
								},
								&parse.ResultColumnExpression{
									Expression: &parse.ExpressionColumn{
										Column: "username",
										Table:  "u",
									},
									Alias: "author",
								},
							},
							From: &parse.RelationTable{
								Table: "posts",
								Alias: "p",
							},
							Joins: []*parse.Join{
								{
									Type: parse.JoinTypeInner,
									Relation: &parse.RelationTable{
										Table: "users",
										Alias: "u",
									},
									On: &parse.ExpressionComparison{
										Left: &parse.ExpressionColumn{
											Column: "author_id",
											Table:  "p",
										},
										Operator: parse.ComparisonOperatorEqual,
										Right: &parse.ExpressionColumn{
											Column: "id",
											Table:  "u",
										},
									},
								},
							},
							Where: &parse.ExpressionComparison{
								Left: &parse.ExpressionColumn{
									Column: "username",
									Table:  "u",
								},
								Operator: parse.ComparisonOperatorEqual,
								Right: &parse.ExpressionLiteral{
									Type:  types.TextType,
									Value: "satoshi",
								},
							},
						},
					},

					Ordering: []*parse.OrderingTerm{
						{
							Expression: &parse.ExpressionColumn{
								Table:  "u",
								Column: "username",
							},
							Order: parse.OrderTypeDesc,
							Nulls: parse.NullOrderLast,
						},
						// apply default ordering
						{
							Expression: &parse.ExpressionColumn{
								Table:  "p",
								Column: "id",
							},
						},
						{
							Expression: &parse.ExpressionColumn{
								Table:  "u",
								Column: "id",
							},
						},
					},
				},
			},
		},
		{
			name: "delete",
			sql:  "delete from users where id = 1;",
			want: &parse.SQLStatement{
				SQL: &parse.DeleteStatement{
					Table: "users",
					Where: &parse.ExpressionComparison{
						Left: &parse.ExpressionColumn{
							Column: "id",
						},
						Operator: parse.ComparisonOperatorEqual,
						Right: &parse.ExpressionLiteral{
							Type:  types.IntType,
							Value: int64(1),
						},
					},
				},
			},
		},
		{
			name: "upsert with conflict - success",
			sql:  `INSERT INTO users (id) VALUES (1) ON CONFLICT (id) DO UPDATE SET id = users.id + excluded.id;`,
			want: &parse.SQLStatement{
				SQL: &parse.InsertStatement{
					Table:   "users",
					Columns: []string{"id"},
					Values: [][]parse.Expression{
						{
							&parse.ExpressionLiteral{
								Type:  types.IntType,
								Value: int64(1),
							},
						},
					},
					Upsert: &parse.UpsertClause{
						ConflictColumns: []string{"id"},
						DoUpdate: []*parse.UpdateSetClause{
							{
								Column: "id",
								Value: &parse.ExpressionArithmetic{
									Left: &parse.ExpressionColumn{
										Column: "id",
										Table:  "users",
									},
									Operator: parse.ArithmeticOperatorAdd,
									Right: &parse.ExpressionColumn{
										Column: "id",
										Table:  "excluded",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "upsert with conflict - ambiguous error",
			sql:  `INSERT INTO users (id) VALUES (1) ON CONFLICT (id) DO UPDATE SET id = id + 1;`,
			err:  parse.ErrAmbiguousConflictTable,
		},
		{
			name: "select against unnamed procedure",
			sql:  "select * from get_all_user_ids();",
			want: &parse.SQLStatement{
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnWildcard{},
							},
							From: &parse.RelationFunctionCall{
								FunctionCall: &parse.ExpressionFunctionCall{
									Name: "get_all_user_ids",
								},
							},
						},
					},
					// no ordering since the procedure implementation is ordered
				},
			},
		},
		{
			name: "select join with unnamed subquery",
			sql: `SELECT p.id as id, u.username as author FROM posts AS p
			INNER JOIN (SELECT id as uid FROM users WHERE id = 1) ON p.author_id = uid;`,
			err: parse.ErrUnnamedJoin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parse.ParseSQL(tt.sql, &types.Schema{
				Name: "mydb",
				Tables: []*types.Table{
					tblUsers,
					tblPosts,
				},
				Procedures: []*types.Procedure{
					procGetAllUserIds,
				},
			})
			require.NoError(t, err)

			if res.ParseErrs.Err() != nil {
				if tt.err == nil {
					t.Errorf("unexpected error: %v", res.ParseErrs.Err())
				} else {
					require.ErrorIs(t, res.ParseErrs.Err(), tt.err)
				}

				return
			}

			if !deepCompare(res.AST, tt.want) {
				t.Errorf("unexpected AST:\n%s", diff(res.AST, tt.want))
			}
		})
	}
}

// deepCompare deep compares the values of two nodes.
// It ignores the parseTypes.Node field.
func deepCompare(node1, node2 any) bool {
	// we return true for the parseTypes.Node field,
	// we also need to ignore the unexported "schema" fields
	return cmp.Equal(node1, node2, cmpOpts()...)
}

// diff returns the diff between two nodes.
func diff(node1, node2 any) string {
	return cmp.Diff(node1, node2, cmpOpts()...)
}

func cmpOpts() []cmp.Option {
	return []cmp.Option{
		cmp.AllowUnexported(
			parse.ExpressionLiteral{},
			parse.ExpressionFunctionCall{},
			parse.ExpressionForeignCall{},
			parse.ExpressionVariable{},
			parse.ExpressionArrayAccess{},
			parse.ExpressionMakeArray{},
			parse.ExpressionFieldAccess{},
			parse.ExpressionParenthesized{},
			parse.ExpressionColumn{},
			parse.ExpressionSubquery{},
			parse.ProcedureStmtDeclaration{},
			parse.ProcedureStmtAssign{},
			parse.ProcedureStmtCall{},
			parse.ProcedureStmtForLoop{},
			parse.ProcedureStmtIf{},
			parse.ProcedureStmtSQL{},
			parse.ProcedureStmtBreak{},
			parse.ProcedureStmtReturn{},
			parse.ProcedureStmtReturnNext{},
			parse.LoopTermRange{},
			parse.LoopTermSQL{},
			parse.LoopTermVariable{},
		),
		cmp.Comparer(func(x, y parse.Position) bool {
			return true
		}),
	}
}
