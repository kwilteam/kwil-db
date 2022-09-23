package postgres

import "github.com/kwilteam/kwil-db/internal/schemadef/schema"

func tname(t *schema.Table) (string, string) {
	if t.Schema != nil {
		return t.Schema.Name, t.Name
	}
	return "", t.Name
}
