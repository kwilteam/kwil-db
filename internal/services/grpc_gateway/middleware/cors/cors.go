package cors

import (
	"net/http"

	"github.com/kwilteam/kwil-db/internal/services/grpc_gateway/middleware"
)

const (
	AllowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	AllowHeaders = "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType, X-Api-Key"
)

func allowedOrigin(cors []string, origin string) bool {
	return true
	/*
		if len(cors) == 1 && cors[0] == "*" {
			return true
		}
		// allow multiple origins
		for _, s := range cors {
			if matched, _ := regexp.MatchString(s, origin); matched {
				return true
			}
		}
		return false
	*/
}

func MCors(cors []string) *middleware.NamedMiddleware {
	return &middleware.NamedMiddleware{
		Name: "cors",
		Middleware: func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if allowedOrigin(cors, r.Header.Get("Origin")) {
					w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
					w.Header().Set("Access-Control-Allow-Methods", AllowMethods)
					w.Header().Set("Access-Control-Allow-Headers", AllowHeaders)
				}

				if r.Method == "OPTIONS" {
					return
				}

				h.ServeHTTP(w, r)
			})
		},
	}
}
