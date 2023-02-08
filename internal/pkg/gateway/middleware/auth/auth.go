package auth

import (
	"context"
	"go.uber.org/zap"
	"kwil/internal/pkg/gateway/middleware"
	"kwil/pkg/log"
	"net/http"
)

const ApiKeyHeader = "X-Api-Key"
const MessageUnauthorized = `{"message": "request unauthorized"}`

type User struct{}

func setUser(r *http.Request, u *User) *http.Request {
	ctxWithUser := context.WithValue(r.Context(), "userkey", u)
	return r.WithContext(ctxWithUser)
}

func MAuth(m Manager, logger log.Logger) *middleware.NamedMiddleware {
	logger = logger.Named("auth")
	return &middleware.NamedMiddleware{
		Name: "auth",
		Middleware: func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				apiKey := r.Header.Get(ApiKeyHeader)
				t := &token{ApiKey: apiKey}
				if !m.IsAllowed(t) {
					logger.Info("request unauthorized", zap.String("api_key", apiKey))
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusUnauthorized)
					_, err := w.Write([]byte(MessageUnauthorized))
					if err != nil {
						logger.Error("failed to write response", zap.Error(err))
					}
					return
				}

				r = setUser(r, nil)
				h.ServeHTTP(w, r)
			})
		},
	}
}
