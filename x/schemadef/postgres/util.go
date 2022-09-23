package postgres

import "kwil/x/schemadef/schema"

func tname(t *schema.Table) (string, string) {
	if t.Schema != nil {
		return t.Schema.Name, t.Name
	}
	return "", t.Name
}
