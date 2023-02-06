package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vektah/gqlparser/gqlerror"
	"go.uber.org/zap"
	"io"
	"kwil/internal/pkg/graphql/hasura"
	"kwil/internal/pkg/graphql/misc"
	"kwil/pkg/log"
	log2 "log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type graphqlReq struct {
	Query string `json:"query"`
}

type RProxy struct {
	logger log.Logger
	proxy  *httputil.ReverseProxy
}

func NewRProxy(cfg Config, logger log.Logger) *RProxy {
	ru, err := url.Parse(cfg.Endpoint)
	if err != nil {
		log2.Fatal(err)
	}

	u := ru.JoinPath("v1")

	logger.Info("graphql endpoint configured", zap.String("endpoint", u.String()))
	go hasura.Initialize(cfg.Endpoint, logger)

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
