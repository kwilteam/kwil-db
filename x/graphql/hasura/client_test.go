package hasura

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
)

func init() {
	viper.BindEnv(AdminSecretName, AdminSecretEnv)
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
		w.Write(jsonBody)

	}))
	defer backendServer.Close()

	url := backendServer.URL

	h := NewClient(url)

	err := h.TrackTable(DefaultSource, DefaultSchema, "table1")
	if err != nil {
		t.Errorf("trackTable() should success, err=%v", err)
	}

	err = h.TrackTable(DefaultSource, DefaultSchema, "table1")
	if err == nil {
		t.Errorf("trackTable() should raise err to track a already tracked table")
	}
}
