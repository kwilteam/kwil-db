package swagger

import (
	"bytes"
	_ "embed"
	v0 "kwil/api/openapi-spec/api/v0"
	"net/http"
	"time"
)

func GWSwaggerJSONHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	var t time.Time
	http.ServeContent(w, r, "swagger.json", t, bytes.NewReader(v0.Swagger))
}

func GWSwaggerUIHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	var t time.Time
	http.ServeContent(w, r, "index.html", t, bytes.NewReader(v0.SwaggerUI))
}
