package parse_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/order"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_Kuneiform tests the Kuneiform parser.
func Test_Kuneiform(t *testing.T) {
	type testCase struct {
		name string
		kf   string
		want *types.Schema
		err  error // can be nil
		// checkAfterErr will continue with the schema comparison
		// after an error is encountered.
		checkAfterErr bool
	}

	tests := []testCase{
		{
			name: "simple schema",
			kf: `
			database mydb;

			table users {
				id int primary_key notnull,
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
				id int primary not null,
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
				id int pk,
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
		{
			name: "two database blocks",
			kf: `database a;
			database b;`,
			err: parse.ErrSyntax,
		},
		{
			// tests for https://github.com/kwilteam/kwil-db/issues/752
			name:          "incomplete database block",
			kf:            `datab`,
			want:          &types.Schema{},
			err:           parse.ErrSyntax,
			checkAfterErr: true,
		},
		{
			// similar to the above test, the same edge case existed for foreign procedures
			name: "incomplete foreign procedure",
			kf: `database a;
			foreign proce`,
			want: &types.Schema{
				Name: "a",
				ForeignProcedures: []*types.ForeignProcedure{
					{}, // there will be one empty foreign procedure
				},
			},
			err:           parse.ErrSyntax,
			checkAfterErr: true,
		},
		{
			// this test tests for properly handling errors for missing primary keys
			name: "missing primary key",
			kf: `
			database mydb;

			table users {
				id int not null
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
										Type: types.NOT_NULL,
									},
								},
							},
						},
					},
				},
			},
			err: parse.ErrNoPrimaryKey,
		},
		{
			name: "empty body",
			kf: `
			database mydb;

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
					},
				},
			},
		},
		{
			name: "foreign key to non-existent table",
			kf: `database mydb;

			table a {
				id int primary key,
				foreign key (id) references b (id)
			}
			`,
			err: parse.ErrUnknownTable,
		},
		{
			name: "foreign key to non-existent column",
			kf: `database mydb;

			table a {
				id int primary key
			}

			table b {
				id int primary key,
				foreign key (id) references a (id_not_exist)
			}
			`,
			err: parse.ErrUnknownColumn,
		},
		{
			name: "foreign key on non-existent column",
			kf: `database mydb;

			table a {
				id int primary key
			}

			table b {
				id int primary key,
				foreign key (id_not_exist) references a (id)
			}
			`,
			err: parse.ErrUnknownColumn,
		},
		{
			name: "index on non-existent column",
			kf: `database mydb;

			table a {
				id int primary key,
				#idx index(id_not_exist)
			}
			`,
			err: parse.ErrUnknownColumn,
		},
		{
			// regression test for https://github.com/kwilteam/kwil-db/issues/896#issue-2423754035
			name: "invalid foreign key",
			kf: `database glow;

			table data {
				id uuid primary key,
				owner_id uuid notnull,
				foreign key (owner_id) references users(id) on update cascade
				// TODO: add other columns
			}
			`,
			err: parse.ErrUnknownTable,
		},
		{
			// regression test for https://github.com/kwilteam/kwil-db/issues/896#issue-2423754035
			name: "invalid foreign key",
			kf: `database mydb;

			table a {
				id int primary key
			}

			table b {
				id int primary key,
				id2 int,
				foreign key (id2) references a(id2)
			}
			`,
			err: parse.ErrUnknownColumn,
		},
		{
			// regression test for https://github.com/kwilteam/kwil-db/issues/896#issuecomment-2243806123
			name: "max on non-numeric type",
			kf: `database mydb;

			table a {
				id uuid primary key,
				age text max(100)
			}
			`,
			err: parse.ErrColumnConstraint,
		},
		{
			// regression test for https://github.com/kwilteam/kwil-db/issues/896#issuecomment-2243835819
			name: "mex_len on blob",
			kf: `database mydb;

			table a {
				id uuid primary key,
				bts blob maxlen(100)
			}
			`,
			want: &types.Schema{
				Name: "mydb",
				Tables: []*types.Table{
					{
						Name: "a",
						Columns: []*types.Column{
							{
								Name: "id",
								Type: types.UUIDType,
								Attributes: []*types.Attribute{
									{
										Type: types.PRIMARY_KEY,
									},
								},
							},
							{
								Name: "bts",
								Type: types.BlobType,
								Attributes: []*types.Attribute{
									{
										Type:  types.MAX_LENGTH,
										Value: "100",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "conflict with function",
			kf: `database mydb;

			table a {
				id int primary key,
				age int max(100)
			}

			procedure max() public view {}
			`,
			err: parse.ErrReservedKeyword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parse.ParseSchemaWithoutValidation([]byte(tt.kf))
			require.NoError(t, err)
			if tt.err != nil {
				parseErrs := res.ParseErrs.Errors()
				if len(parseErrs) == 0 {
					require.Fail(t, "expected parse errors")
				}

				require.ErrorIs(t, parseErrs[0], tt.err)
				if !tt.checkAfterErr {
					return
				}
			} else {
				require.NoError(t, res.ParseErrs.Err())
				if tt.checkAfterErr {
					panic("cannot use checkAfterErr without an error")
				}
			}

			assertPositionsAreSet(t, res.ParsedActions)
			assertPositionsAreSet(t, res.ParsedProcedures)

			require.EqualValues(t, tt.want, res.Schema)

			// we will also test that the schemas were properly cleaned.
			// we test this by copying the schema to a new schema, cleaning the new schema, and comparing the two.
			bts, err := json.Marshal(res.Schema)
			require.NoError(t, err)

			var got2 types.Schema
			err = json.Unmarshal(bts, &got2)
			require.NoError(t, err)

			// since checkAfterErr means we expect a parser error, we shouldn't clean since
			// it will likely fail since the schema is invalid anyways
			if !tt.checkAfterErr {
				err = got2.Clean()
				require.NoError(t, err)
			}

			got2.Owner = nil // unmarshal sets Owner to empty array, so we need to set it to nil to compare

			require.EqualValues(t, res.Schema, &got2)
		})
	}
}

// assertPositionsAreSet asserts that all positions in the ast are set.
func assertPositionsAreSet(t *testing.T, v any) {
	parse.RecursivelyVisitPositions(v, func(gp parse.GetPositioner) {
		pos := gp.GetPosition()
		// if not set, this will tell us the struct
		assert.True(t, pos.IsSet, "position is not set. struct type: %T", gp)
	})
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

	foreignProcGetUser = &types.ForeignProcedure{
		Name: "get_user_id",
		Parameters: []*types.DataType{
			types.TextType,
		},
		Returns: &types.ProcedureReturn{
			Fields: []*types.NamedType{
				{
					Name: "id",
					Type: types.IntType,
				},
			},
		},
	}

	foreignProcCreateUser = &types.ForeignProcedure{
		Name: "foreign_create_user",
		Parameters: []*types.DataType{
			types.IntType,
			types.TextType,
		},
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
			name: "procedure applies default ordering to selects",
			proc: `
			select * from users;
			`,
			want: &parse.ProcedureParseResult{
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtSQL{
						SQL: &parse.SQLStatement{
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
								Ordering: []*parse.OrderingTerm{
									{
										Expression: exprColumn("users", "id"),
									},
								},
							},
						},
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
			$arr6 := $arr5[1:2];
			$arr7 := $arr5[1:];
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
					"$arr6": types.ArrayType(types.IntType),
					"$arr7": types.ArrayType(types.IntType),
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
					&parse.ProcedureStmtAssign{
						Variable: exprVar("$arr6"),
						Value: &parse.ExpressionArrayAccess{
							Array: exprVar("$arr5"),
							FromTo: [2]parse.Expression{
								exprLit(1),
								exprLit(2),
							},
						},
					},
					&parse.ProcedureStmtAssign{
						Variable: exprVar("$arr7"),
						Value: &parse.ExpressionArrayAccess{
							Array: exprVar("$arr5"),
							FromTo: [2]parse.Expression{
								exprLit(1),
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
		{
			name: "missing return values",
			returns: &types.ProcedureReturn{
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
			proc: `return 1;`,
			err:  parse.ErrReturn,
		},
		{
			name: "no return values",
			returns: &types.ProcedureReturn{
				Fields: []*types.NamedType{{
					Name: "id",
					Type: types.IntType,
				}},
			},
			proc: `$a := 1;`,
			err:  parse.ErrReturn,
		},
		{
			name: "if/then missing return in one branch",
			returns: &types.ProcedureReturn{
				Fields: []*types.NamedType{{
					Name: "id",
					Type: types.IntType,
				}},
			},
			proc: `
			if true {
				return 1;
			} else {
				$a := 1;
			}
			`,
			err: parse.ErrReturn,
		},
		{
			name: "for loop with if return",
			returns: &types.ProcedureReturn{
				Fields: []*types.NamedType{{
					Name: "id",
					Type: types.IntType,
				}},
			},
			proc: `
			$arr := [1,2,3];
			for $i in $arr {
				if $i == -1 {
					break;
				}
				return $i;
			}
			`,
			err: parse.ErrReturn,
		},
		{
			name: "nested for loop",
			returns: &types.ProcedureReturn{
				Fields: []*types.NamedType{{
					Name: "id",
					Type: types.IntType,
				}},
			},
			proc: `
			$arr int[];
			for $i in $arr {
				for $j in 1..$i {
					break; // only breaks the inner loop
				}

				return $i; // this will always exit on first $i iteration
			}
			`,
			want: &parse.ProcedureParseResult{
				Variables: map[string]*types.DataType{
					"$arr": types.ArrayType(types.IntType),
					"$i":   types.IntType,
					"$j":   types.IntType,
				},
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtDeclaration{
						Variable: exprVar("$arr"),
						Type:     types.ArrayType(types.IntType),
					},
					&parse.ProcedureStmtForLoop{
						Receiver: exprVar("$i"),
						LoopTerm: &parse.LoopTermVariable{
							Variable: exprVar("$arr"),
						},
						Body: []parse.ProcedureStmt{
							&parse.ProcedureStmtForLoop{
								Receiver: exprVar("$j"),
								LoopTerm: &parse.LoopTermRange{
									Start: exprLit(1),
									End:   exprVar("$i"),
								},
								Body: []parse.ProcedureStmt{
									&parse.ProcedureStmtBreak{},
								},
							},
							&parse.ProcedureStmtReturn{
								Values: []parse.Expression{exprVar("$i")},
							},
						},
					},
				},
			},
		},
		{
			name: "returns table incorrect",
			returns: &types.ProcedureReturn{
				Fields: []*types.NamedType{{
					Name: "id",
					Type: types.IntType,
				}},
			},
			proc: `return select id from users;`, // this is intentional- plpgsql treats this as a table return
			err:  parse.ErrReturn,
		},
		{
			name: "returns table correct",
			returns: &types.ProcedureReturn{
				IsTable: true,
				Fields: []*types.NamedType{{
					Name: "id",
					Type: types.IntType,
				}},
			},
			proc: `return select 1 as id;`,
			want: &parse.ProcedureParseResult{
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtReturn{
						SQL: &parse.SQLStatement{
							SQL: &parse.SelectStatement{
								SelectCores: []*parse.SelectCore{
									{
										Columns: []parse.ResultColumn{
											&parse.ResultColumnExpression{
												Expression: exprLit(1),
												Alias:      "id",
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
			name: "returns next incorrect",
			returns: &types.ProcedureReturn{
				Fields: []*types.NamedType{{
					Name: "id",
					Type: types.IntType,
				}},
			},
			proc: `$a int[];
			for $row in $a {
				return next $row;
			}
			`,
			err: parse.ErrReturn,
		},
		{
			name: "returns next correct",
			returns: &types.ProcedureReturn{
				IsTable: true,
				Fields: []*types.NamedType{{
					Name: "id",
					Type: types.IntType,
				}},
			},
			proc: `
			for $row in select * from get_all_user_ids() {
				return next $row.id;
			}
			`,
			want: &parse.ProcedureParseResult{
				CompoundVariables: map[string]struct{}{
					"$row": {},
				},
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtForLoop{
						Receiver: exprVar("$row"),
						LoopTerm: &parse.LoopTermSQL{
							Statement: &parse.SQLStatement{
								SQL: &parse.SelectStatement{
									SelectCores: []*parse.SelectCore{
										{
											Columns: []parse.ResultColumn{&parse.ResultColumnWildcard{}},
											From: &parse.RelationFunctionCall{
												FunctionCall: &parse.ExpressionFunctionCall{
													Name: "get_all_user_ids",
												},
											},
										},
									},
									Ordering: []*parse.OrderingTerm{
										{
											Expression: exprColumn("", "id"),
										},
									},
								},
							},
						},
						Body: []parse.ProcedureStmt{
							&parse.ProcedureStmtReturnNext{
								Values: []parse.Expression{
									&parse.ExpressionFieldAccess{
										Record: exprVar("$row"),
										Field:  "id",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "error func exits",
			proc: `error('error message');`,
			returns: &types.ProcedureReturn{
				Fields: []*types.NamedType{{
					Name: "id",
					Type: types.IntType,
				}},
			},
			want: &parse.ProcedureParseResult{
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtCall{
						Call: &parse.ExpressionFunctionCall{
							Name: "error",
							Args: []parse.Expression{
								exprLit("error message"),
							},
						},
					},
				},
			},
		},
		{
			// this tests for regression on a previously known bug
			name: "foreign procedure returning nothing to a variable",
			returns: &types.ProcedureReturn{
				Fields: []*types.NamedType{
					{
						Name: "id",
						Type: types.IntType,
					},
				},
			},
			proc: `
			return foreign_create_user['xbd', 'create_user'](1, 'user1');
			`,
			err: parse.ErrType,
		},
		{
			// regression test for a previously known bug
			name: "calling a procedure that returns nothing works fine",
			proc: `
			foreign_create_user['xbd', 'create_user'](1, 'user1');
			`,
			want: &parse.ProcedureParseResult{
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtCall{
						Call: &parse.ExpressionForeignCall{
							Name: "foreign_create_user",
							ContextualArgs: []parse.Expression{
								exprLit("xbd"),
								exprLit("create_user"),
							},
							Args: []parse.Expression{
								exprLit(1),
								exprLit("user1"),
							},
						},
					},
				},
			},
		},
		{
			// this is a regression test for a previous bug
			name: "discarding return values of a function is ok",
			proc: `abs(-1);`,
			want: &parse.ProcedureParseResult{
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtCall{
						Call: &parse.ExpressionFunctionCall{
							Name: "abs",
							Args: []parse.Expression{
								exprLit(-1),
							},
						},
					},
				},
			},
		},
		{
			name: "sum types - failure",
			proc: `
			$sum := 0;
			for $row in select sum(id) as id from users {
				$sum := $sum + $row.id;
			}
			`,
			// this should error, since sum returns numeric
			err: parse.ErrType,
		},
		{
			name: "sum types - success",
			proc: `
			$sum decimal(1000,0);
			for $row in select sum(id) as id from users {
				$sum := $sum + $row.id;
			}
			`,
			want: &parse.ProcedureParseResult{
				Variables: map[string]*types.DataType{
					"$sum": mustNewDecimal(1000, 0),
				},
				CompoundVariables: map[string]struct{}{
					"$row": {},
				},
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtDeclaration{
						Variable: exprVar("$sum"),
						Type:     mustNewDecimal(1000, 0),
					},
					&parse.ProcedureStmtForLoop{
						Receiver: exprVar("$row"),
						LoopTerm: &parse.LoopTermSQL{
							Statement: &parse.SQLStatement{
								SQL: &parse.SelectStatement{
									SelectCores: []*parse.SelectCore{
										{
											Columns: []parse.ResultColumn{
												&parse.ResultColumnExpression{
													Expression: &parse.ExpressionFunctionCall{
														Name: "sum",
														Args: []parse.Expression{
															exprColumn("", "id"),
														},
													},
													Alias: "id",
												},
											},
											From: &parse.RelationTable{
												Table: "users",
											},
										},
									},
									// If there is an aggregate clause with no group by, then no ordering is applied.
								},
							},
						},
						Body: []parse.ProcedureStmt{
							&parse.ProcedureStmtAssign{
								Variable: exprVar("$sum"),
								Value: &parse.ExpressionArithmetic{
									Left:     exprVar("$sum"),
									Operator: parse.ArithmeticOperatorAdd,
									Right:    &parse.ExpressionFieldAccess{Record: exprVar("$row"), Field: "id"},
								},
							},
						},
					},
				},
			},
		},
		{
			// this is a regression test for a previous bug
			name: "adding arrays",
			proc: `
			$arr1 := [1,2,3];
			$arr2 := [4,5,6];
			$arr3 := $arr1 + $arr2;
			`,
			err: parse.ErrType,
		},
		{
			// this is a regression test for a previous bug
			name: "early return",
			proc: `
						return;
						`,
			want: &parse.ProcedureParseResult{
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtReturn{},
				},
			},
		},
		{
			name: "variable scoping",
			proc: `
			for $i in 1..10 {
			}
			$j := $i;
			`,
			err: parse.ErrUndeclaredVariable,
		},
		{
			name: "scoping 2",
			proc: `
			for $i in 1..10 {
				$j := $i;
			}
			$k := $j;
			`,
			err: parse.ErrUndeclaredVariable,
		},
		{
			name: "if scoping",
			proc: `
			if true {
				$i := 1;
			}
			$j := $i;
			`,
			err: parse.ErrUndeclaredVariable,
		},
		{
			name: "else scoping",
			proc: `
			if false {
			} else {
				$i := 1;
			}
			$j := $i;
		`,
			err: parse.ErrUndeclaredVariable,
		},
		{ // regression test
			name: "equals order of operations",
			proc: `
			$a := 1+2 == 3;
			`,
			want: &parse.ProcedureParseResult{
				Variables: map[string]*types.DataType{
					"$a": types.BoolType,
				},
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtAssign{
						Variable: exprVar("$a"),
						Value: &parse.ExpressionComparison{
							Left:     &parse.ExpressionArithmetic{Left: exprLit(1), Operator: parse.ArithmeticOperatorAdd, Right: exprLit(2)},
							Operator: parse.ComparisonOperatorEqual,
							Right:    exprLit(3),
						},
					},
				},
			},
		},
		{
			// regression test https://github.com/kwilteam/kwil-db/pull/947
			name: "string literal",
			proc: `
			$a := '\'hello\'';
			`,
			want: &parse.ProcedureParseResult{
				Variables: map[string]*types.DataType{
					"$a": types.TextType,
				},
				AST: []parse.ProcedureStmt{
					&parse.ProcedureStmtAssign{
						Variable: exprVar("$a"),
						Value:    exprLit("\\'hello\\'"),
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
				ForeignProcedures: []*types.ForeignProcedure{
					foreignProcGetUser,
					foreignProcCreateUser,
				},
			})
			require.NoError(t, err)

			if tt.err != nil {
				require.ErrorIs(t, res.ParseErrs.Err(), tt.err)
				return
			}
			require.NoError(t, res.ParseErrs.Err())

			assertPositionsAreSet(t, res.AST)

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

func mustNewDecimal(precision, scale uint16) *types.DataType {
	dt, err := types.NewDecimalType(precision, scale)
	if err != nil {
		panic(err)
	}
	return dt
}

// exprLit makes an ExpressionLiteral.
// it can only make strings and ints
func exprLit(v any) parse.Expression {
	switch t := v.(type) {
	case int:
		isNeg := t < 0
		if isNeg {
			t *= -1
		}

		liter := &parse.ExpressionLiteral{
			Type:  types.IntType,
			Value: int64(t),
			Typecastable: parse.Typecastable{
				TypeCast: types.IntType,
			},
		}

		if isNeg {
			return &parse.ExpressionUnary{
				Operator:   parse.UnaryOperatorNeg,
				Expression: liter,
			}
		}

		return liter
	case int64:
		isNeg := t < 0
		if isNeg {
			t *= -1
		}

		liter := &parse.ExpressionLiteral{
			Type:  types.IntType,
			Value: t,
			Typecastable: parse.Typecastable{
				TypeCast: types.IntType,
			},
		}

		if isNeg {
			return &parse.ExpressionUnary{
				Operator:   parse.UnaryOperatorNeg,
				Expression: liter,
			}
		}

		return liter
	case string:
		return &parse.ExpressionLiteral{
			Type:  types.TextType,
			Value: t,
			Typecastable: parse.Typecastable{
				TypeCast: types.TextType,
			},
		}
	case bool:
		return &parse.ExpressionLiteral{
			Type:  types.BoolType,
			Value: t,
			Typecastable: parse.Typecastable{
				TypeCast: types.BoolType,
			},
		}
	default:
		panic("TEST ERROR: invalid type for literal")
	}
}

func exprFunctionCall(name string, args ...parse.Expression) *parse.ExpressionFunctionCall {
	return &parse.ExpressionFunctionCall{
		Name: name,
		Args: args,
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
									Expression: exprColumn("", "id"),
									Alias:      "i",
								},
								&parse.ResultColumnExpression{
									Expression: &parse.ExpressionFunctionCall{
										Name: "length",
										Args: []parse.Expression{
											exprColumn("", "username"),
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
								Left:     exprColumn("u", "id"),
								Operator: parse.ComparisonOperatorEqual,
								Right:    exprLit(1),
							},
						},
					},
					// apply default ordering
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("u", "id"),
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
							exprLit(1),
							exprLit(1),
						},
						{
							exprLit(2),
							&parse.ExpressionSubquery{
								Subquery: &parse.SelectStatement{
									SelectCores: []*parse.SelectCore{
										{
											Columns: []parse.ResultColumn{
												&parse.ResultColumnExpression{
													Expression: exprColumn("", "id"),
												},
											},
											From: &parse.RelationTable{
												Table: "users",
											},
											Where: &parse.ExpressionComparison{
												Left:     exprColumn("", "username"),
												Operator: parse.ComparisonOperatorEqual,
												Right:    exprLit("user2"),
											},
										},
									},
									Limit: exprLit(1),
									// apply default ordering
									Ordering: []*parse.OrderingTerm{
										{
											Expression: exprColumn("users", "id"),
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
									Expression: exprColumn("p", "id"),
									Alias:      "id",
								},
								&parse.ResultColumnExpression{
									Expression: exprColumn("u", "username"),
									Alias:      "author",
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
										Left:     exprColumn("p", "author_id"),
										Operator: parse.ComparisonOperatorEqual,
										Right:    exprColumn("u", "id"),
									},
								},
							},
							Where: &parse.ExpressionComparison{
								Left:     exprColumn("u", "username"),
								Operator: parse.ComparisonOperatorEqual,
								Right:    exprLit("satoshi"),
							},
						},
					},

					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("u", "username"),
							Order:      parse.OrderTypeDesc,
							Nulls:      parse.NullOrderLast,
						},
						// apply default ordering
						{
							Expression: exprColumn("p", "id"),
						},
						{
							Expression: exprColumn("u", "id"),
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
						Left:     exprColumn("", "id"),
						Operator: parse.ComparisonOperatorEqual,
						Right:    exprLit(1),
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
							exprLit(1),
						},
					},
					Upsert: &parse.UpsertClause{
						ConflictColumns: []string{"id"},
						DoUpdate: []*parse.UpdateSetClause{
							{
								Column: "id",
								Value: &parse.ExpressionArithmetic{
									Left:     exprColumn("users", "id"),
									Operator: parse.ArithmeticOperatorAdd,
									Right:    exprColumn("excluded", "id"),
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
					// apply default ordering
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("", "id"),
						},
					},
				},
			},
		},
		{
			name: "select join with unnamed subquery",
			sql: `SELECT p.id as id, u.username as author FROM posts AS p
			INNER JOIN (SELECT id as uid FROM users WHERE id = 1) ON p.author_id = uid;`,
			err: parse.ErrUnnamedJoin,
		},
		{
			name: "compound select",
			sql:  `SELECT * FROM users union SELECT * FROM users;`,
			want: &parse.SQLStatement{
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
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnWildcard{},
							},
							From: &parse.RelationTable{
								Table: "users",
							},
						},
					},
					CompoundOperators: []parse.CompoundOperator{
						parse.CompoundOperatorUnion,
					},
					// apply default ordering
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("", "id"),
						},
						{
							Expression: exprColumn("", "username"),
						},
					},
				},
			},
		},
		{
			name: "compound with mismatched shape",
			sql:  `SELECT username, id FROM users union SELECT id, username FROM users;`,
			err:  parse.ErrResultShape,
		},
		{
			name: "compound selecting 1 column",
			sql:  `SELECT username FROM users union SELECT username FROM users;`,
			want: &parse.SQLStatement{
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnExpression{
									Expression: exprColumn("", "username"),
								},
							},
							From: &parse.RelationTable{
								Table: "users",
							},
						},
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnExpression{
									Expression: exprColumn("", "username"),
								},
							},
							From: &parse.RelationTable{
								Table: "users",
							},
						},
					},
					CompoundOperators: []parse.CompoundOperator{parse.CompoundOperatorUnion},
					// apply default ordering
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("", "username"),
						},
					},
				},
			},
		},
		{
			name: "group by",
			sql:  `SELECT u.username, count(u.id) FROM users as u GROUP BY u.username HAVING count(u.id) > 1;`,
			want: &parse.SQLStatement{
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnExpression{
									Expression: exprColumn("u", "username"),
								},
								&parse.ResultColumnExpression{
									Expression: &parse.ExpressionFunctionCall{
										Name: "count",
										Args: []parse.Expression{
											exprColumn("u", "id"),
										},
									},
								},
							},
							From: &parse.RelationTable{
								Table: "users",
								Alias: "u",
							},
							GroupBy: []parse.Expression{
								exprColumn("u", "username"),
							},
							Having: &parse.ExpressionComparison{
								Left: &parse.ExpressionFunctionCall{
									Name: "count",
									Args: []parse.Expression{
										exprColumn("u", "id"),
									},
								},
								Operator: parse.ComparisonOperatorGreaterThan,
								Right:    exprLit(1),
							},
						},
					},
					// apply default ordering
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("u", "username"),
						},
					},
				},
			},
		},
		{
			name: "group by with having, having is in group by clause",
			// there's a much easier way to write this query, but this is just to test the parser
			sql: `SELECT username FROM users GROUP BY username HAVING length(username) > 1;`,
			want: &parse.SQLStatement{
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnExpression{
									Expression: exprColumn("", "username"),
								},
							},
							From: &parse.RelationTable{
								Table: "users",
							},
							GroupBy: []parse.Expression{
								exprColumn("", "username"),
							},
							Having: &parse.ExpressionComparison{
								Left: &parse.ExpressionFunctionCall{
									Name: "length",
									Args: []parse.Expression{
										exprColumn("", "username"),
									},
								},
								Operator: parse.ComparisonOperatorGreaterThan,
								Right:    exprLit(1),
							},
						},
					},
					// apply default ordering
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("users", "username"),
						},
					},
				},
			},
		},
		{
			name: "group by with having, invalid column without aggregate",
			sql:  `SELECT u.username, count(u.id) FROM users as u GROUP BY u.username HAVING u.id > 1;`,
			err:  parse.ErrAggregate,
		},
		{
			name: "compound select with group by",
			sql:  `SELECT u.username, count(u.id) FROM users as u GROUP BY u.username HAVING count(u.id) > 1 UNION SELECT u.username, count(u.id) FROM users as u GROUP BY u.username HAVING count(u.id) > 1;`,
			err:  parse.ErrAggregate,
		},
		{
			name: "aggregate with no group by returns many columns",
			sql:  `SELECT count(u.id), u.username FROM users as u;`,
			err:  parse.ErrAggregate,
		},
		{
			name: "aggregate with no group by returns one column",
			sql:  `SELECT count(*) FROM users;`,
			want: &parse.SQLStatement{
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnExpression{
									Expression: &parse.ExpressionFunctionCall{
										Name: "count",
										Star: true,
									},
								},
							},
							From: &parse.RelationTable{
								Table: "users",
							},
						},
					},
				},
			},
		},
		{
			name: "aggregate with no group by and ordering fails",
			sql:  `SELECT count(*) FROM users order by count(*) DESC;`,
			err:  parse.ErrAggregate,
		},
		{
			name: "ordering for subqueries",
			sql:  `SELECT u.username, p.id FROM (SELECT * FROM users) as u inner join (SELECT * FROM posts) as p on u.id = p.author_id;`,
			want: &parse.SQLStatement{
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnExpression{
									Expression: exprColumn("u", "username"),
								},
								&parse.ResultColumnExpression{
									Expression: exprColumn("p", "id"),
								},
							},
							From: &parse.RelationSubquery{
								Subquery: &parse.SelectStatement{
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
									Ordering: []*parse.OrderingTerm{
										{
											Expression: exprColumn("users", "id"),
										},
									},
								},
								Alias: "u",
							},
							Joins: []*parse.Join{
								{
									Type: parse.JoinTypeInner,
									Relation: &parse.RelationSubquery{
										Subquery: &parse.SelectStatement{
											SelectCores: []*parse.SelectCore{
												{
													Columns: []parse.ResultColumn{
														&parse.ResultColumnWildcard{},
													},
													From: &parse.RelationTable{
														Table: "posts",
													},
												},
											},
											Ordering: []*parse.OrderingTerm{
												{
													Expression: exprColumn("posts", "id"),
												},
											},
										},
										Alias: "p",
									},
									On: &parse.ExpressionComparison{
										Left:     exprColumn("u", "id"),
										Operator: parse.ComparisonOperatorEqual,
										Right:    exprColumn("p", "author_id"),
									},
								},
							},
						},
					},
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("u", "id"),
						},
						{
							Expression: exprColumn("u", "username"),
						},
						{
							Expression: exprColumn("p", "id"),
						},
						{
							Expression: exprColumn("p", "author_id"),
						},
					},
				},
			},
		},
		{
			name: "select against subquery with table join",
			sql:  `SELECT u.username, p.id FROM (SELECT * FROM users) inner join posts as p on users.id = p.author_id;`,
			err:  parse.ErrUnnamedJoin,
		},
		{
			name: "default ordering on procedure call",
			sql:  `SELECT * FROM get_all_user_ids();`,
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
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("", "id"),
						},
					},
				},
			},
		},
		{
			name: "join against unnamed function call fails",
			sql:  `SELECT * FROM users inner join get_all_user_ids() on users.id = u.id;`,
			err:  parse.ErrUnnamedJoin,
		},
		{name: "non utf-8", sql: "\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98;", err: parse.ErrSyntax},
		{
			// this select doesn't make much sense, however
			// it is a regression test for a previously known bug
			// https://github.com/kwilteam/kwil-db/pull/810
			name: "offset and limit",
			sql:  `SELECT * FROM users LIMIT id OFFSET id;`,
			want: &parse.SQLStatement{
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
					Offset: exprColumn("", "id"),
					Limit:  exprColumn("", "id"),
					// apply default ordering
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("users", "id"),
						},
					},
				},
			},
		},
		{
			// this is a regression test for a previous bug.
			// when parsing just SQL, we can have unknown variables
			name: "unknown variable is ok",
			sql:  `SELECT $id;`,
			want: &parse.SQLStatement{
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnExpression{
									Expression: exprVar("$id"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "select JOIN, with no specified INNER/OUTER",
			sql: `SELECT u.* FROM users as u
			JOIN posts as p ON u.id = p.author_id;`, // default is INNER
			want: &parse.SQLStatement{
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnWildcard{
									Table: "u",
								},
							},
							From: &parse.RelationTable{
								Table: "users",
								Alias: "u",
							},
							Joins: []*parse.Join{
								{
									Type: parse.JoinTypeInner,
									Relation: &parse.RelationTable{
										Table: "posts",
										Alias: "p",
									},
									On: &parse.ExpressionComparison{
										Left:     exprColumn("u", "id"),
										Operator: parse.ComparisonOperatorEqual,
										Right:    exprColumn("p", "author_id"),
									},
								},
							},
						},
					},
					// apply default ordering
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("u", "id"),
						},
						{
							Expression: exprColumn("p", "id"),
						},
					},
				},
			},
		},
		{
			// regression tests for a previous bug, where whitespace after
			// the semicolon would cause the parser to add an extra semicolon
			name: "whitespace after semicolon",
			sql:  "SELECT 1;     ",
			want: &parse.SQLStatement{
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnExpression{
									Expression: exprLit(1),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "cte",
			sql:  `WITH cte AS (SELECT id FROM users) SELECT * FROM cte;`,
			want: &parse.SQLStatement{
				CTEs: []*parse.CommonTableExpression{
					{
						Name: "cte",
						Query: &parse.SelectStatement{
							SelectCores: []*parse.SelectCore{
								{
									Columns: []parse.ResultColumn{
										&parse.ResultColumnExpression{
											Expression: exprColumn("", "id"),
										},
									},
									From: &parse.RelationTable{
										Table: "users",
									},
								},
							},
							// apply default ordering
							Ordering: []*parse.OrderingTerm{
								{
									Expression: exprColumn("users", "id"),
								},
							},
						},
					},
				},
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnWildcard{},
							},
							From: &parse.RelationTable{
								Table: "cte",
							},
						},
					},
					// apply default ordering
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("cte", "id"),
						},
					},
				},
			},
		},
		{
			name: "cte with columns",
			sql:  `WITH cte (id2) AS (SELECT id FROM users) SELECT * FROM cte;`,
			want: &parse.SQLStatement{
				CTEs: []*parse.CommonTableExpression{
					{
						Name:    "cte",
						Columns: []string{"id2"},
						Query: &parse.SelectStatement{
							SelectCores: []*parse.SelectCore{
								{
									Columns: []parse.ResultColumn{
										&parse.ResultColumnExpression{
											Expression: exprColumn("", "id"),
										},
									},
									From: &parse.RelationTable{
										Table: "users",
									},
								},
							},
							// apply default ordering
							Ordering: []*parse.OrderingTerm{
								{
									Expression: exprColumn("users", "id"),
								},
							},
						},
					},
				},
				SQL: &parse.SelectStatement{
					SelectCores: []*parse.SelectCore{
						{
							Columns: []parse.ResultColumn{
								&parse.ResultColumnWildcard{},
							},
							From: &parse.RelationTable{
								Table: "cte",
							},
						},
					},
					// apply default ordering
					Ordering: []*parse.OrderingTerm{
						{
							Expression: exprColumn("cte", "id2"),
						},
					},
				},
			},
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
			}, false)
			require.NoError(t, err)

			if res.ParseErrs.Err() != nil {
				if tt.err == nil {
					t.Errorf("unexpected error: %v", res.ParseErrs.Err())
				} else {
					require.ErrorIs(t, res.ParseErrs.Err(), tt.err)
				}

				return
			}

			assertPositionsAreSet(t, res.AST)

			if !deepCompare(tt.want, res.AST) {
				t.Errorf("unexpected AST:%s", diff(tt.want, res.AST))
			}
		})
	}
}

func exprColumn(t, c string) *parse.ExpressionColumn {
	return &parse.ExpressionColumn{
		Table:  t,
		Column: c,
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

func Test_Actions(t *testing.T) {
	type testcase struct {
		name   string
		tables []*types.Table
		action *types.Action
		want   *parse.ActionParseResult
		err    error
	}

	tests := []testcase{
		{
			name:   "return value",
			tables: []*types.Table{tableBalances},
			action: &types.Action{
				Name:   "check_balance",
				Public: false,
				Modifiers: []types.Modifier{
					types.ModifierView,
				},
				Body: "SELECT        CASE            WHEN balance < 10 THEN ERROR('insufficient balance')            ELSE null        END    FROM balances WHERE wallet = @caller;",
			},
			want: &parse.ActionParseResult{
				AST: []parse.ActionStmt{
					&parse.ActionStmtSQL{
						SQL: &parse.SQLStatement{
							SQL: &parse.SelectStatement{
								SelectCores: []*parse.SelectCore{
									{
										Columns: []parse.ResultColumn{
											&parse.ResultColumnExpression{
												Expression: &parse.ExpressionCase{
													WhenThen: [][2]parse.Expression{
														{
															&parse.ExpressionComparison{
																Left:     exprColumn("", "balance"),
																Operator: parse.ComparisonOperatorLessThan,
																Right:    exprLit(10),
															},
															exprFunctionCall("error", exprLit("insufficient balance")),
														},
													},
													Else: &parse.ExpressionLiteral{
														Value: nil,
														Type:  types.NullType,
													},
												},
											},
										},
										From: &parse.RelationTable{
											Table: "balances",
										},
										Where: &parse.ExpressionComparison{
											Left:     exprColumn("", "wallet"),
											Operator: parse.ComparisonOperatorEqual,
											Right:    exprVar("@caller"),
										},
									},
								},
								Ordering: []*parse.OrderingTerm{
									{
										Expression: exprColumn("balances", "wallet"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "action in-line statement calls select",
			action: &types.Action{
				Name:       "check_balance",
				Parameters: []string{"$arg"},
				Body:       "$res = my_ext.my_method($arg[0]);",
			},
			err: parse.ErrAssignment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parse.ParseAction(tt.action, &types.Schema{
				Name:   "mydb",
				Tables: tt.tables,
				Procedures: []*types.Procedure{
					procGetAllUserIds,
				},
			})
			require.NoError(t, err)

			if tt.err != nil {
				require.ErrorIs(t, res.ParseErrs.Err(), tt.err)
				return
			}
			require.NoError(t, res.ParseErrs.Err())
			res.ParseErrs = nil

			assertPositionsAreSet(t, res.AST)

			if !deepCompare(tt.want, res) {
				t.Errorf("unexpected output: %s", diff(tt.want, res))
			}
		})
	}
}

var tableBalances = &types.Table{
	Name: "balances",
	Columns: []*types.Column{
		{
			Name: "wallet",
			Type: types.TextType,
			Attributes: []*types.Attribute{
				{
					Type: types.PRIMARY_KEY,
				},
			},
		},
		{
			Name: "balance",
			Type: types.IntType,
		},
	},
}

// this tests full end-to-end parsing of a schema, with full validation.
// It is mostly necessary to test for bugs that slip through the cracks
// of testing individual components.
func Test_FullParse(t *testing.T) {
	type testcase struct {
		name string
		kf   string
		err  error // if nil, no error is expected
	}

	tests := []testcase{
		{
			// this is a regression test for a previous bug where the parser would panic
			// when a procedure had a return statement with no body.
			name: "empty body with returns",
			kf: `database proxy;
// admin simply tracks the schema admins
table admin {
  address text primary key
}
// add_admin adds a new admin to the schema.
// only current admins can add new admins
procedure add_admin ($address text) public  {}
procedure is_admin ($address text) public view returns (bool) {}
			`,
			err: parse.ErrReturn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parse.Parse([]byte(tt.kf))
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
