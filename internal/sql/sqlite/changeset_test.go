package sqlite_test

import (
	"context"
	"encoding/hex"
	"testing"

	sql "github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
	"github.com/stretchr/testify/require"
)

func Test_Changeset(t *testing.T) {
	type testCase struct {
		name string
		// initial statements will seed the database before the changeset is generated
		initial []string
		// stmts are the statements that will be executed to generate the changeset
		stmts   []string
		results *sqlite.Changeset
	}

	testCases := []testCase{
		{
			name: "insert",
			stmts: []string{
				"INSERT INTO users (id, name, age) VALUES (1, 'alice', 42)",
				"INSERT INTO users (id, name, age) VALUES (2, 'bob', 43)",
			},
			results: &sqlite.Changeset{
				Tables: map[string]*sqlite.TableChangeset{
					"users": {
						ColumnNames: []string{"id", "name", "age"},
						Records: map[string]*sqlite.RecordChange{
							pkId(1): {
								ChangeType: sqlite.RecordChangeTypeCreate,
								Values: []*sqlite.Value{
									sqlVal(1),
									sqlVal("alice"),
									sqlVal(42),
								},
							},
							pkId(2): {
								ChangeType: sqlite.RecordChangeTypeCreate,
								Values: []*sqlite.Value{
									sqlVal(2),
									sqlVal("bob"),
									sqlVal(43),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "update",
			initial: []string{
				"INSERT INTO users (id, name, age) VALUES (1, 'alice', 42)",
				"INSERT INTO users (id, name, age) VALUES (2, 'bob', 43)",
			},
			stmts: []string{
				"UPDATE users SET name = 'alice2' WHERE id = 1",
				"UPDATE users SET name = 'bob2', age = 44 WHERE id = 2",
			},
			results: &sqlite.Changeset{
				Tables: map[string]*sqlite.TableChangeset{
					"users": {
						ColumnNames: []string{"id", "name", "age"},
						Records: map[string]*sqlite.RecordChange{
							pkId(1): {
								ChangeType: sqlite.RecordChangeTypeUpdate,
								Values:     vals(nil, "alice2", nil),
							},
							pkId(2): {
								ChangeType: sqlite.RecordChangeTypeUpdate,
								Values:     vals(nil, "bob2", 44),
							},
						},
					},
				},
			},
		},
		{
			name: "delete",
			initial: []string{
				"INSERT INTO users (id, name, age) VALUES (1, 'alice', 42)",
				"INSERT INTO users (id, name, age) VALUES (2, 'bob', 43)",
			},
			stmts: []string{
				"DELETE FROM users",
			},
			results: &sqlite.Changeset{
				Tables: map[string]*sqlite.TableChangeset{
					"users": {
						ColumnNames: []string{"id", "name", "age"},
						Records: map[string]*sqlite.RecordChange{
							pkId(1): {
								ChangeType: sqlite.RecordChangeTypeDelete,
							},
							pkId(2): {
								ChangeType: sqlite.RecordChangeTypeDelete,
							},
						},
					},
				},
			},
		},
		{
			name: "insert and delete",
			stmts: []string{
				"INSERT INTO users (id, name, age) VALUES (1, 'alice', 42)",
				"DELETE FROM users",
			},
			results: &sqlite.Changeset{
				Tables: map[string]*sqlite.TableChangeset{},
			},
		},
		{
			name: "insert and update",
			stmts: []string{
				"INSERT INTO users (id, name, age) VALUES (1, 'alice', 42)",
				"UPDATE users SET name = 'alice2' WHERE id = 1",
			},
			results: &sqlite.Changeset{
				Tables: map[string]*sqlite.TableChangeset{
					"users": {
						ColumnNames: []string{"id", "name", "age"},
						Records: map[string]*sqlite.RecordChange{
							pkId(1): {
								ChangeType: sqlite.RecordChangeTypeCreate,
								Values:     []*sqlite.Value{sqlVal(1), sqlVal("alice2"), sqlVal(42)},
							},
						},
					},
				},
			},
		},
	}

	// tests the changeset is deterministic
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			conn, err := openDB("test", sql.OpenCreate)
			require.NoError(t, err)
			defer deleteTempDir()
			defer func() {
				err = conn.Close()
				require.NoError(t, err)
			}()

			err = createUserTable(ctx, conn)
			require.NoError(t, err)

			for _, stmt := range tc.initial {
				res, err := conn.Execute(ctx, stmt, nil)
				require.NoError(t, err)
				err = res.Finish()
				require.NoError(t, err)
			}

			session, err := conn.CreateSession()
			require.NoError(t, err)
			defer session.Delete()

			for _, stmt := range tc.stmts {
				res, err := conn.Execute(ctx, stmt, nil)
				require.NoError(t, err)
				err = res.Finish()
				require.NoError(t, err)
			}

			sqliteSession, ok := session.(*sqlite.Session)
			require.True(t, ok) // doing this to test that the chageset returns the correct results

			changeset, err := sqliteSession.Changeset(ctx)
			require.NoError(t, err)

			require.EqualValues(t, tc.results, changeset)
		})

		// tests the ID() method is deterministic
		t.Run(tc.name+" ID determinism", func(t *testing.T) {
			ids := [][]byte{}

			for i := 0; i < 10; i++ {
				func() {
					ctx := context.Background()
					conn, err := openDB("test", sql.OpenCreate)
					require.NoError(t, err)
					defer deleteTempDir()
					defer func() {
						err = conn.Close()
						require.NoError(t, err)
					}()

					err = createUserTable(ctx, conn)
					require.NoError(t, err)

					for _, stmt := range tc.initial {
						res, err := conn.Execute(ctx, stmt, nil)
						require.NoError(t, err)
						err = res.Finish()
						require.NoError(t, err)
					}

					session, err := conn.CreateSession()
					require.NoError(t, err)
					defer session.Delete()

					for _, stmt := range tc.stmts {
						res, err := conn.Execute(ctx, stmt, nil)
						require.NoError(t, err)
						err = res.Finish()
						require.NoError(t, err)
					}

					id, err := session.ChangesetID(ctx)
					require.NoError(t, err)

					ids = append(ids, id)
				}()
			}

			for i := 1; i < len(ids)-1; i++ {
				require.Equal(t, ids[i], ids[i+1])
			}
		})
	}
}

// pkId converts the values to sqlite values and marshals them
func pkId(v ...any) string {

	res, err := sqlite.ValueSet(vals(v...)).MarshalBinary()
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(res)
}

// vals converts the values to sqlite values
func vals(v ...any) []*sqlite.Value {
	var result []*sqlite.Value
	for _, val := range v {
		result = append(result, sqlVal(val))
	}

	return result
}

func sqlVal(v any) *sqlite.Value {
	var dataType sqlite.DataType
	var assertedVal any

	switch val := v.(type) {
	case nil:
		dataType = sqlite.DataTypeNull
		assertedVal = nil
	case int:
		dataType = sqlite.DataTypeInt
		assertedVal = int64(val)
	case string:
		dataType = sqlite.DataTypeText
		assertedVal = val
	default:
		panic("unsupported test type")
	}

	return &sqlite.Value{
		DataType: dataType,
		Value:    assertedVal,
	}
}
