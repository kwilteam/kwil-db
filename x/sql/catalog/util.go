package catalog

func sameType(a, b *QualName) bool {
	// The pg_catalog schema is searched by default, so take that into
	// account when comparing schemas
	aSchema := a.Schema
	bSchema := b.Schema
	if aSchema == "pg_catalog" {
		aSchema = ""
	}
	if bSchema == "pg_catalog" {
		bSchema = ""
	}
	if aSchema != bSchema {
		return false
	}
	if a.Name != b.Name {
		return false
	}
	return true
}
