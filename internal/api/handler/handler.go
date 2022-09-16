package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kwilteam/kwil-db/internal/logx"
	"go.uber.org/zap"
)

type Handler struct {
	Router *mux.Router
	Server *http.Server
	Auth   PeerAuthenticator
	Logger logx.Logger
}

type PeerAuthenticator interface {
	Authenticate(*websocket.Conn) error
}

func NewHandler(logger logx.Logger, a PeerAuthenticator) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
		Auth:   a,
		Logger: logger,
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
		h.Logger.Error("failed to upgrade to websocket", zap.Error(err))
		return
	}
	defer conn.Close()

	err = h.Auth.Authenticate(conn)
	if err != nil {
		h.Logger.Error("failed to authenticate", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}
