package api

import _ "embed"

// xxx go:embed v0.swagger.json
//var SwaggerV0 []byte

//go:embed v1.swagger.json
var SwaggerV1 []byte

//go:embed swaggerui.html
var SwaggerUI []byte
