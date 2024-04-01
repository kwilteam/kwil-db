package testkit

import (
	"fmt"

	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

type mockCatalog struct {
	tables map[string]*dt.Schema
}

func (m *mockCatalog) GetDataSource(tableRef *dt.TableRef) (ds.DataSource, error) {
	relName := tableRef.Table // for testing, we ignore the database name
	schema, ok := m.tables[relName]
	if !ok {
		return nil, fmt.Errorf("table %s not found", relName)
	}
	return ds.NewExampleDataSource(schema), nil
}

func InitMockCatalog() *mockCatalog {
	stubUserData, _ := ds.NewCSVDataSource("../testdata/users.csv")
	stubPostData, _ := ds.NewCSVDataSource("../testdata/posts.csv")
	commentsData, _ := ds.NewCSVDataSource("../testdata/comments.csv")
	commentRelData, _ := ds.NewCSVDataSource("../testdata/comment_rel.csv")

	return &mockCatalog{
		tables: map[string]*dt.Schema{
			// for testing, we ignore the database name
			"users":       stubUserData.Schema(),
			"posts":       stubPostData.Schema(),
			"comments":    commentsData.Schema(),
			"comment_rel": commentRelData.Schema(),
		},
	}
}
