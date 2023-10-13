package swagger

import (
	"bytes"
	_ "embed"
	"net/http"
	"time"

	swagger "github.com/kwilteam/kwil-db/internal/services/http/api"
)

// func GWSwaggerJSONV0Handler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
// 	var t time.Time
// 	http.ServeContent(w, r, "swagger.json", t, bytes.NewReader(swagger.SwaggerV0))
// }

func GWSwaggerJSONV1Handler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	var t time.Time
	http.ServeContent(w, r, "swagger.json", t, bytes.NewReader(swagger.SwaggerV1))
}

func GWSwaggerUIHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	var t time.Time
	http.ServeContent(w, r, "index.html", t, bytes.NewReader(swagger.SwaggerUI))
}
