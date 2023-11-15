package swagger

import (
	"bytes"
	_ "embed"
	"net/http"
	"time"

	swagger "github.com/kwilteam/kwil-db/internal/services/http/api"
)

func GWSwaggerJSONTxV1Handler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	var t time.Time
	http.ServeContent(w, r, "tx.swagger.json", t, bytes.NewReader(swagger.SwaggerTxV1))
}

func GWSwaggerUIHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	var t time.Time
	http.ServeContent(w, r, "index.html", t, bytes.NewReader(swagger.SwaggerUI))
}
