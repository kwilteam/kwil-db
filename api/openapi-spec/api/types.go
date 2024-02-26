package v0

import _ "embed"

//go:embed v0/api.swagger.json
var SwaggerV0 []byte

//go:embed v1/api.swagger.json
var SwaggerV1 []byte

//go:embed swaggerui.html
var SwaggerUI []byte
