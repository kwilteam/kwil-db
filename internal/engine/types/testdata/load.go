package testdata

import (
	_ "embed"
)

//go:embed testdata.json
var schemaFile []byte
