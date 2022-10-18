package apipb

import (
	"bytes"
	_ "embed"
	"net/http"
	"time"
)

//go:embed api.swagger.json
var swagger []byte

func ServeSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	var t time.Time
	http.ServeContent(w, r, "swagger.json", t, bytes.NewReader(swagger))
}

//go:embed swaggerui.html
var swaggerUI []byte

func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	var t time.Time
	http.ServeContent(w, r, "index.html", t, bytes.NewReader(swaggerUI))
}
