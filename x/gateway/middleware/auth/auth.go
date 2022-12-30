package auth

import (
	"context"
	"kwil/x/gateway/middleware"
	"net/http"
)

const ApiKeyHeader = "X-Api-Key"

type User struct{}

type Auth struct {
	h http.Handler
	m Manager
}

func newAuth(h http.Handler, m Manager) *Auth {
	return &Auth{h: h, m: m}
}

// setUser pass the auth user for future usage
func (a *Auth) setUser(r *http.Request, u *User) *http.Request {
	ctxWithUser := context.WithValue(r.Context(), "userkey", u)
	return r.WithContext(ctxWithUser)
}
func (a *Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get(ApiKeyHeader)
	t := &token{ApiKey: apiKey}
	if a.m.IsAllowed(t) {
		r = a.setUser(r, nil)
		a.h.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(""))
	}
}

func MAuth(m Manager) middleware.Middleware {
	return func(h http.Handler) http.Handler {
		return newAuth(h, m)
	}
}
