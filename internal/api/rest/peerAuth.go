package rest

import (
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Accepting all requests
	},
}

func (h *Handler) PeerAuth(w http.ResponseWriter, r *http.Request) {
	// upgrade to websocket
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
