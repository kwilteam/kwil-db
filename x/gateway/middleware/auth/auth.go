package auth

import (
	"context"
	"kwil/x/gateway/middleware"
	"net/http"
)

const ApiKeyHeader = "X-Api-Key"
const MessageUnauthorized = `{"message": "request unauthorized"}`

type User struct{}

func setUser(r *http.Request, u *User) *http.Request {
	ctxWithUser := context.WithValue(r.Context(), "userkey", u)
	return r.WithContext(ctxWithUser)
}

func MAuth(m Manager) *middleware.NamedMiddleware {
	return &middleware.NamedMiddleware{
		Name: "auth",
		Middleware: func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				apiKey := r.Header.Get(ApiKeyHeader)
				t := &token{ApiKey: apiKey}
				if !m.IsAllowed(t) {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(MessageUnauthorized))
					return
				}

				r = setUser(r, nil)
				h.ServeHTTP(w, r)
			})
		},
	}
}
