package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"kwil/x/logx"
)

type PeerAuthenticator interface {
	Authenticate(*websocket.Conn) error
}

func New(logger logx.Logger, authenticator PeerAuthenticator) http.Handler {
	serveMux := mux.NewRouter()
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Accepting all requests
		},
	}

	serveMux.HandleFunc("/api/v0/peer-auth", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("failed to upgrade to websocket", zap.Error(err))
			return
		}
		defer conn.Close()

		err = authenticator.Authenticate(conn)
		if err != nil {
			logger.Error("failed to authenticate", zap.Error(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	return serveMux
}
