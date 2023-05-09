package ddlbuilder

import (
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"strings"
)

type builder interface {
	Build() string
}

type schemaPicker interface {
	Table(string) constraintPicker
	Schema(string) tablePicker
}

type tablePicker interface {
	Table(string) constraintPicker
}

func generateName(schema, table, column, typ string) string {
	sb := &strings.Builder{}
	sb.WriteString(schema)
	sb.WriteByte('_') // adding delimiter to avoid potential collisions
	sb.WriteString(table)
	sb.WriteByte('_')
	sb.WriteString(column)
	sb.WriteByte('_')
	sb.WriteString(typ)

	hash := crypto.Sha256Hex([]byte(sb.String()))
	return hash[:63]
}
