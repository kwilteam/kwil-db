package hasura

import (
	"github.com/kwilteam/kwil-db/pkg/log"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/spf13/viper"
)

func init() {
	_ = viper.BindEnv(AdminSecretFlag, AdminSecretEnv)
}

func Test_trackTable(t *testing.T) {
	tracked := false
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jsonBody []byte
		w.Header().Set("Content-type", "application/json")

		if tracked {
			jsonBody = []byte(`{"code":"already-tracked","error":"view/table already tracked: \"test\"","path":"$.args"}`)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			jsonBody = []byte(`{"message": "success"}`)
			w.WriteHeader(http.StatusOK)
		}
		tracked = !tracked
		_, _ = w.Write(jsonBody)
	}))
	defer backendServer.Close()
	url := backendServer.URL
	logger, _ := zap.NewDevelopment()
	h := NewClient(url, log.Logger{L: logger})
	err := h.TrackTable(DefaultSource, DefaultSchema, "table1")
	if err != nil {
		t.Errorf("trackTable() should success, err=%v", err)
	}

	err = h.TrackTable(DefaultSource, DefaultSchema, "table1")
	if err == nil {
		t.Errorf("trackTable() should raise err to track a already tracked table")
	}
}
