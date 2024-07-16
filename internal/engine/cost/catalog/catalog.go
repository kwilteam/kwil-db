package catalog

import (
	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

type Catalog interface {
	GetDataSource(tableRef *dt.TableRef) (ds.DataSource, error)
}

type defaultCatalogProvider struct {
	dbidAliases map[string]string // alias -> dbid
	srcs        map[string]ds.DataSource
}
