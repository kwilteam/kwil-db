package specifications

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"time"
)

var (
	SchemaLoader DatabaseSchemaLoader = &FileDatabaseSchemaLoader{
		Modifier: func(db *transactions.Schema) {
			// NOTE: this is a hack to make sure the db name is temporary unique
			db.Name = fmt.Sprintf("%s_%s", db.Name, time.Now().Format("20160102"))
		}}
)

func SetSchemaLoader(loader DatabaseSchemaLoader) {
	SchemaLoader = loader
}
