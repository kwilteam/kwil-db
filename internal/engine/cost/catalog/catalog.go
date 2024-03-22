package catalog

import (
	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

type Catalog interface {
	GetSchemaSource(tableRef *dt.TableRef) (ds.SchemaSource, error)
}

type defaultCatalogProvider struct {
	dbidAliases map[string]string // alias -> dbid
	schemas     map[string]ds.SchemaSource
}
