package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	Router *mux.Router
	Server *http.Server
	Auth   Authenticator
}

type Authenticator interface {
	Authenticate(*websocket.Conn) error
}

func NewHandler(a Authenticator) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
		Auth:   a,
	}

	h.Router.HandleFunc("/api/v0/peer-auth", h.PeerAuth)

	h.Server = &http.Server{
		Addr:    ":8080",
		Handler: h.Router,
	}

	return h
}

func (h *Handler) PeerAuth(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Accepting all requests
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to upgrade to websocket")
		return
	}
	defer conn.Close()

	err = h.Auth.Authenticate(conn)
	if err != nil {
		log.Error().Err(err).Msg("failed to authenticate")
		return
	}

	w.WriteHeader(http.StatusOK)
}
