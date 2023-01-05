package apisvc

import (
	"net/http"

	"kwil/x/logx"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func NewHandler(logger logx.Logger) http.Handler {
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

		// commenting this out since I'm removing auth for now
		// it won't be in the first version and it will likely change a lot
		/*err = authenticator.Authenticate(conn)
		if err != nil {
			logger.Error("failed to authenticate", zap.Error(err))
			return
		}*/

		w.WriteHeader(http.StatusOK)
	})

	return serveMux
}
