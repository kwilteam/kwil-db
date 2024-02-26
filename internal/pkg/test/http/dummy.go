package http

import (
	"fmt"
	"net/http"
)

type DummyHttpHandler struct {
	Data string
}

func (d DummyHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, d.Data)
}
