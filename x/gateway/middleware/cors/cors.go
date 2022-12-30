package cors

import (
	"kwil/x/gateway/middleware"
	"net/http"
	"regexp"
	"strings"
)

const (
	AllowMethods    = "GET, POST, PATCH, DELETE"
	AllowHeaders    = "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType, X-Api-Key"
	GatewayCorsName = "gateway-cors"
	GatewayCorsEnv  = "GATEWAY_CORS"
)

type Cors struct {
	h    http.Handler
	cors string
}

func newCors(h http.Handler, cors string) *Cors {
	return &Cors{h: h, cors: cors}
}

func (c *Cors) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if allowedOrigin(c.cors, r.Header.Get("Origin")) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", AllowMethods)
		w.Header().Set("Access-Control-Allow-Headers", AllowHeaders)
	}

	if r.Method == "OPTIONS" {
		return
	}

	c.h.ServeHTTP(w, r)
}

func MCors(cors string) *middleware.NamedMiddleware {
	return &middleware.NamedMiddleware{
		Name: "cors",
		Mw: func(h http.Handler) http.Handler {
			return newCors(h, cors)
		},
	}
}

func allowedOrigin(cors, origin string) bool {
	if cors == "*" {
		return true
	}
	// allow multiple origins
	for _, s := range strings.Split(cors, ",") {
		if matched, _ := regexp.MatchString(s, origin); matched {
			return true
		}
	}
	return false
}
