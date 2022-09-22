package postgres

import "github.com/kwilteam/kwil-db/internal/hcl"

// TypeRegistry contains the supported TypeSpecs for the Postgres driver.
var TypeRegistry = hcl.NewRegistry(
	hcl.WithParser(ParseType),
	hcl.WithSpecs(
		hcl.NewTypeSpec("string", hcl.WithAttributes(hcl.SizeTypeAttr(false))),
		hcl.NewTypeSpec("int"),
		hcl.NewTypeSpec("bool"),
		hcl.NewTypeSpec("date"),
		hcl.NewTypeSpec("datetime"),
	),
)
