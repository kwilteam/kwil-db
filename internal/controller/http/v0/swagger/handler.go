package swagger

import (
	_ "embed"
)

//// @yaiba TODO: locate this file
////go:embed api.swagger.json
//var swagger []byte
//
//func GWSwaggerJSONHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
//	var t time.Time
//	http.ServeContent(w, r, "swagger.json", t, bytes.NewReader(swagger))
//}
//
////go:embed swaggerui.html
//var swaggerUI []byte
//
//func GWSwaggerUIHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
//	var t time.Time
//	http.ServeContent(w, r, "index.html", t, bytes.NewReader(swaggerUI))
//}
