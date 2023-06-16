package serialize

import (
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset2"
)

type SchemaMetadata struct {
}

func SerializeProcedure(procedure *dataset2.Procedure) ([]byte, error) {
	return fmt.Sprintf("action %s", action.Name)
}

func SerializeExtension(extension *dataset2.Extension) ([]byte, error) {
	return fmt.Sprintf("extension %s", name)
}
