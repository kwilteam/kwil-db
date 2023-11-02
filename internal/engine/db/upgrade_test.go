package db_test

import (
	"context"
	"io"
	"testing"

	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/internal/engine/db"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/stretchr/testify/assert"
)

// TODO: test storing old versions and upgrading

func Test_UpgradingProcedures(t *testing.T) {

	type testCase struct {
		name              string
		storedProcedure   versionedProcedure
		returnedProcedure *types.Procedure
	}

	testCases := []testCase{
		{
			name: "stored v1",
			storedProcedure: versionedProcedure{
				Version: 1,
				Procedure: &types.Procedure{
					Name: "test_v1",
					Args: []string{
						"$arg1",
						"$arg2",
					},
					Statements: []string{
						"SELECT * FROM users WHERE id = $arg1",
					},
					Public: false,
				},
			},
			returnedProcedure: &types.Procedure{
				Name: "test_v1",
				Args: []string{
					"$arg1",
					"$arg2",
				},
				Statements: []string{
					"SELECT * FROM users WHERE id = $arg1",
				},
				Public: true,
				Modifiers: []types.Modifier{
					types.ModifierOwner,
				},
			},
		},
		{
			name: "stored v2",
			storedProcedure: versionedProcedure{
				Version: 2,
				Procedure: &types.Procedure{
					Name: "test_v1",
					Args: []string{
						"$arg1",
						"$arg2",
					},
					Statements: []string{
						"SELECT * FROM users WHERE id = $arg1",
					},
					Public: false,
					Modifiers: []types.Modifier{
						types.ModifierAuthenticated,
					},
					Annotations: []string{"sql(engine=sqlite3)"},
				},
			},
			returnedProcedure: &types.Procedure{
				Name: "test_v1",
				Args: []string{
					"$arg1",
					"$arg2",
				},
				Statements: []string{
					"SELECT * FROM users WHERE id = $arg1",
				},
				Public: false,
				Modifiers: []types.Modifier{
					types.ModifierAuthenticated,
				},
				Annotations: []string{"sql(engine=sqlite3)"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			datastore := &procedureStore{
				procedures: []*versionedProcedure{
					&tc.storedProcedure,
				},
			}

			ctx := context.Background()
			database, err := db.NewDB(ctx, datastore)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			returnedProcedures, err := database.ListProcedures(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(returnedProcedures) != 1 {
				t.Fatalf("expected 1 procedure, got %v", len(returnedProcedures))
			}

			returned := returnedProcedures[0]

			assert.Equal(t, *tc.returnedProcedure, *returned)
		})
	}
}

// we only need to implement Query
type procedureStore struct {
	procedures []*versionedProcedure
	baseMockDatastore
}

func (m procedureStore) Query(ctx context.Context, query string, args map[string]interface{}) ([]map[string]interface{}, error) {
	/*
		needs to return map[string]any{
			"identifier": "procedure_name",
			"content":    json.Marshal(&db.VersionMetadata{Version: int64(version), Data: json.Marshal(&types.Prrocedure{})}),
		}
	*/

	returnVals := []map[string]interface{}{}

	for _, procedure := range m.procedures {
		serializedProc, err := serialize.Encode(procedure.Procedure)
		if err != nil {
			return nil, err
		}

		dbVersionedProcedure := &db.VersionedMetadata{
			Version: procedure.Version,
			Data:    serializedProc,
		}

		contentBytes, err := serialize.Encode(dbVersionedProcedure)
		if err != nil {
			return nil, err
		}

		returnVals = append(returnVals, map[string]interface{}{
			"identifier": procedure.Procedure.Name,
			"content":    contentBytes,
		})
	}

	return returnVals, nil
}

func (m *procedureStore) QueryUnsafe(ctx context.Context, query string, args map[string]any) ([]map[string]any, error) {
	return m.Query(ctx, query, args)
}

func (m *procedureStore) ApplyChangeset(cs io.Reader) error {
	return nil
}

func (m *procedureStore) CreateSession() (sql.Session, error) {
	return nil, nil
}

type versionedProcedure struct {
	Version   uint
	Procedure *types.Procedure
}
