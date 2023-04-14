package swagger

import (
	"bytes"
	_ "embed"
	swagger "kwil/api/openapi-spec/api"
	"net/http"
	"time"
)

func GWSwaggerJSONHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	var t time.Time
	http.ServeContent(w, r, "swagger.json", t, bytes.NewReader(swagger.Swagger))
}

func GWSwaggerUIHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	var t time.Time
	http.ServeContent(w, r, "index.html", t, bytes.NewReader(swagger.SwaggerUI))
}
