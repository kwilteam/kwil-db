package parse_test

import (
	"encoding/json"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
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

// some default tables for testing
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
)
