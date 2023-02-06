package health

import "net/http"

func GWHealthzHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// TODO: check dependency?
	w.WriteHeader(http.StatusOK)
}

func GWReadyzHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	// won't check dependent services
	w.WriteHeader(http.StatusOK)
}
