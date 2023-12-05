package api

import _ "embed"

//go:embed tx/v1.swagger.json
var SwaggerTxV1 []byte

//go:embed swaggerui.html
var SwaggerUI []byte
