package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

type NamedMiddleware struct {
	Name string
	Mw   Middleware
}
