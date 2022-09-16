package v0

import (
	"bytes"
	_ "embed"
	"net/http"
	"time"
)

//go:embed api.swagger.json
var swagger []byte

func ServeSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, "swagger.json", time.Time{}, bytes.NewReader(swagger))
}

//go:embed swaggerui.html
var swaggerUI []byte

func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(swaggerUI))
}
