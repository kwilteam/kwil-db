package graphql

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"github.com/vektah/gqlparser/gqlerror"
	"io"
	"kwil/x/graphql/hasura"
	"kwil/x/logx"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

	go hasura.InitializeHasura()

	proxy := httputil.NewSingleHostReverseProxy(u)

	return &GraphqlRProxy{
		logger: logx.New().Sugar(),
		proxy:  proxy,
	}
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

		// uncomment below to get actual sql to execute
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
