package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"kwil/x/logx"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/spf13/viper"
	"github.com/vektah/gqlparser/gqlerror"
)

type GraphqlRProxy struct {
	logger logx.SugaredLogger
	proxy  *httputil.ReverseProxy
}

func NewGraphqlRProxy() *GraphqlRProxy {
	ru, err := url.Parse(viper.GetString("graphql"))
	if err != nil {
		log.Fatal(err)
	}

	u := ru.JoinPath("v1")
	proxy := httputil.NewSingleHostReverseProxy(u)

	return &GraphqlRProxy{
		logger: logx.New().Sugar(),
		proxy:  proxy,
	}
}

func isMutation(query string) bool {
	// NOTE: enough to correctly block most mutations
	operations := []string{}
	rightBracket := -1
	opens := 0
	for i, c := range query {
		if c == '}' {
			opens -= 1
			if opens == 0 {
				rightBracket = i
			}
		}

		if c == '{' {
			if opens == 0 {
				operations = append(operations, query[rightBracket+1:i])
			}
			opens += 1
		}
	}

	for _, op := range operations {
		if strings.Contains(op, "mutation") {
			return true
		}
	}
	return false
}

func JSONError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(err)
}

func (g *GraphqlRProxy) makeHasuraHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			JSONError(w, fmt.Errorf("parse request failed"), http.StatusInternalServerError)
			g.logger.Errorf("parse request failed: %s", err.Error())
			return
		}

		bodyString := string(bodyBytes)
		if isMutation(bodyString) {
			e := gqlerror.Errorf("Only query is allowed")
			JSONError(w, e, http.StatusBadRequest)
			return
		}

		// // compile GraphQL queries to sql
		// sql, err := g.hasura.ExplainQuery(bodyString)
		// if err != nil {
		// 	JSONError(w, err, http.StatusBadRequest)
		// 	return
		// }
		// // apply ACL to sql

		// restore body
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		fn(w, r)
	}
}

func (g *GraphqlRProxy) Handler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	g.makeHasuraHandler(g.proxy.ServeHTTP)(w, r)
}
