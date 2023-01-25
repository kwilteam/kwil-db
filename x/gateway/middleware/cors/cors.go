package cors

import (
	"kwil/x/gateway/middleware"
	"net/http"
)

const (
	AllowMethods    = "GET, POST, PATCH, DELETE"
	AllowHeaders    = "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType, X-Api-Key"
	GatewayCorsFlag = "gateway-cors"
	GatewayCorsEnv  = "GATEWAY_CORS"
)

func allowedOrigin(cors, origin string) bool {
	return true
	/*
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
	*/
}

func MCors(cors string) *middleware.NamedMiddleware {
	return &middleware.NamedMiddleware{
		Name: "cors",
		Middleware: func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "OPTIONS" {
					if allowedOrigin(cors, r.Header.Get("Origin")) {
						w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
						w.Header().Set("Access-Control-Allow-Methods", AllowMethods)
						w.Header().Set("Access-Control-Allow-Headers", AllowHeaders)
					}
					return
				}

				h.ServeHTTP(w, r)
			})
		},
	}
}
