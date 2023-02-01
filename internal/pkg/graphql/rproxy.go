package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	hasura2 "kwil/internal/pkg/graphql/hasura"
	"kwil/internal/pkg/graphql/misc"
	"kwil/pkg/logger"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/spf13/viper"
	"github.com/vektah/gqlparser/gqlerror"
)

type graphqlReq struct {
	Query string `json:"query"`
}

type RProxy struct {
	logger logger.Logger
	proxy  *httputil.ReverseProxy
}

func NewRProxy() *RProxy {
	ru, err := url.Parse(viper.GetString(hasura2.GraphqlEndpointFlag))
	if err != nil {
		log.Fatal(err)
	}

	u := ru.JoinPath("v1")

	logger := logger.New()
	logger.Info("graphql endpoint configured", zap.String("endpoint", viper.GetString(hasura2.GraphqlEndpointFlag)))
	go hasura2.Initialize()

	proxy := httputil.NewSingleHostReverseProxy(u)

	return &RProxy{
		logger: logger,
		proxy:  proxy,
	}
}

func (g *RProxy) makeHasuraHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			g.logger.Error("parse request failed", zap.Error(err))
			if e := misc.JsonError(w, fmt.Errorf("parse request failed"), http.StatusInternalServerError); e != nil {
				g.logger.Error("write response failed", zap.Error(e))
			}
			return
		}

		var body graphqlReq
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			g.logger.Error("parse request failed", zap.Error(err))
			if e := misc.JsonError(w, fmt.Errorf("parse request failed"), http.StatusBadRequest); e != nil {
				g.logger.Error("write response failed", zap.Error(e))
			}
			return
		}

		if misc.IsMutation(body.Query) {
			err := gqlerror.Errorf("Only query is allowed")
			g.logger.Error("bad request: %s", zap.Error(err))
			if e := misc.JsonError(w, err, http.StatusBadRequest); e != nil {
				g.logger.Error("write response failed", zap.Error(e))
			}
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

func (g *RProxy) Handler(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	g.makeHasuraHandler(g.proxy.ServeHTTP)(w, r)
}
