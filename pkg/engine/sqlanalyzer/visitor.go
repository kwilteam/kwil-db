package sqlanalyzer

import (
	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
)

type SqlVisitor struct {
	tree.BaseVisitor

	metadata *schemaMetadata
}
