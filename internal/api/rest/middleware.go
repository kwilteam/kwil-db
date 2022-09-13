package rest

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kwilteam/kwil-db/internal/chain/config"
	"github.com/rs/zerolog/log"
)

func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msgf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tt, err := strconv.ParseInt(os.Getenv("TIMEOUT_TIME"), 10, 0)
		if err != nil {
			log.Warn().Err(err).Msg("failed to parse timeout time, setting timeout to 20 seconds")
			tt = 20
		}

		t := time.Duration(tt) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), t)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Error().Msgf("panic: %v", r)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func ApiKeyMiddleware(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})
	}
}
