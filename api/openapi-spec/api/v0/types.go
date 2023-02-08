package v0

import _ "embed"

//go:embed api.swagger.json
var Swagger []byte

//go:embed swaggerui.html
var SwaggerUI []byte
