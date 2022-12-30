package cors

import (
	"github.com/spf13/viper"
	"kwil/x/gateway/middleware"
	"net/http"
	"os"
	"regexp"
)

const (
	AllowMethods    = "GET, POST, PATCH, DELETE"
	AllowHeaders    = "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType, X-Api-Key"
	GatewayCorsName = "gateway-cors"
	GatewayCorsEnv  = "GATEWAY_CORS"
)

type Cors struct {
	h http.Handler
}

func newCors(h http.Handler) *Cors {
	return &Cors{h: h}
}

func (c *Cors) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if allowedOrigin(r.Header.Get("Origin")) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", AllowMethods)
		w.Header().Set("Access-Control-Allow-Headers", AllowHeaders)
	}

	if r.Method == "OPTIONS" {
		return
	}

	c.h.ServeHTTP(w, r)
}

func MCors() middleware.Middleware {
	return func(h http.Handler) http.Handler {
		return newCors(h)
	}
}

func allowedOrigin(origin string) bool {
	cors := os.Getenv("GATEWAY_CORS")
	if cors == "" {
		cors = viper.GetString("cors")
	}
	if cors == "*" {
		return true
	}
	if matched, _ := regexp.MatchString(cors, origin); matched {
		return true
	}
	return false
}
