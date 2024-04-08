package specifications

import (
	"fmt"
	"time"

	"github.com/kwilteam/kwil-db/core/types"
)

var (
	defaultTxQueryTimeout = 10 * time.Second
)

var (
	SchemaLoader DatabaseSchemaLoader = &FileDatabaseSchemaLoader{
		Modifier: func(db *types.Schema) {
			// NOTE: this is a hack to make sure the db name is temporary unique
			db.Name = fmt.Sprintf("%s_%s", db.Name, time.Now().Format("20060102"))
		}}
)

func SetSchemaLoader(loader DatabaseSchemaLoader) {
	SchemaLoader = loader
}
