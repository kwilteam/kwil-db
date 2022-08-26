package rest

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/kwilteam/kwil-db/internal/api/service"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	Router  *mux.Router
	Service service.Service
	Server  *http.Server
}

// TODO: Add JSON validation

func NewHandler(service service.Service) *Handler {
	h := &Handler{
		Router:  mux.NewRouter(),
		Service: service,
	}

	h.mapRoutes()
	h.Router.Use(JSONMiddleware)
	h.Router.Use(LoggingMiddleware)
	h.Router.Use(TimeoutMiddleware)

	h.Server = &http.Server{
		Addr:    ":8080",
		Handler: h.Router,
	}

	return h
}

func (h *Handler) Write(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Write")
}

func (h *Handler) mapRoutes() {
	h.Router.HandleFunc("/api/v0/write", JWTAuth(h.Write)).Methods("POST")
	h.Router.HandleFunc("/api/v0/create-database", JWTAuth(h.CreateDatabase)).Methods("POST")
}

func (h *Handler) Serve() error {
	go func() {
		err := h.Server.ListenAndServe()
		if err == nil || err == http.ErrServerClosed {
			return
		}

		log.Fatal().Err(err).Msg("failed to start http server")
		os.Exit(1)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = h.Server.Shutdown(ctx)
	log.Info().Msg("shut down gracefully")

	return nil
}
