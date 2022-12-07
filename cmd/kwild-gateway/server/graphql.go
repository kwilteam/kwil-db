package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/spf13/viper"
	"github.com/vektah/gqlparser/gqlerror"
)

func isMutation(query string) bool {
	// mutation could starts with a operationName or not
	// NOTE: for now this function should be enough
	// TODO: @yaiba a more robust function or remove it
	if strings.Contains(query, "mutation ") || strings.Contains(query, "mutation{") {
		return true
	} else {
		return false
	}
}

func JSONError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(err)
}

func hasuraHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		bodyString := string(bodyBytes)
		if isMutation(bodyString) {
			e := gqlerror.Errorf("Only query is allowed")
			JSONError(w, e, http.StatusBadRequest)
			return
		}

		// restore body
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		fn(w, r)
	}
}

func graphqlHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	ru, err := url.Parse(viper.GetString("graphql"))
	if err != nil {
		log.Fatal(err)
	}

	u := ru.JoinPath("v1")

	proxy := httputil.NewSingleHostReverseProxy(u)
	hasuraHandler(proxy.ServeHTTP)(w, r)
}
