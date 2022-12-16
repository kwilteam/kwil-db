package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"kwil/x/graphql/hasura"
	"kwil/x/graphql/misc"
	"kwil/x/logx"
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
	logger logx.SugaredLogger
	proxy  *httputil.ReverseProxy
}

func NewRProxy() *RProxy {
	ru, err := url.Parse(viper.GetString("graphql"))
	if err != nil {
		log.Fatal(err)
	}

	u := ru.JoinPath("v1")

	go hasura.Initialize()

	proxy := httputil.NewSingleHostReverseProxy(u)

	return &RProxy{
		logger: logx.New().Sugar(),
		proxy:  proxy,
	}
}

func (g *RProxy) makeHasuraHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			g.logger.Errorf("parse request failed: %s", err.Error())
			if e := misc.JsonError(w, fmt.Errorf("parse request failed"), http.StatusInternalServerError); e != nil {
				g.logger.Errorf("write response failed: %s", e.Error())
			}
			return
		}

		var body graphqlReq
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			g.logger.Errorf("parse request failed: %s", err.Error())
			if e := misc.JsonError(w, fmt.Errorf("parse request failed"), http.StatusBadRequest); e != nil {
				g.logger.Errorf("write response failed: %s", e.Error())
			}
			return
		}

		if misc.IsMutation(body.Query) {
			err := gqlerror.Errorf("Only query is allowed")
			g.logger.Errorf("bad request: %s", err.Error())
			if e := misc.JsonError(w, err, http.StatusBadRequest); e != nil {
				g.logger.Errorf("write reponse failed: %s", e.Error())
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
