package health

import (
	"fmt"
	"net/http"
)

func GWHealthzHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// TODO: check dependency?
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

func GWReadyzHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// won't check dependent services
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}
