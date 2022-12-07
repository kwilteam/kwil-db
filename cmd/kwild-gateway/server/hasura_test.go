package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_convertHasuraTableName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "with space",
			args: args{"With Space"},
			want: "with_space",
		},
		{
			name: "without space",
			args: args{"WithoutSpace"},
			want: "withoutspace",
		},
		{
			name: "simple",
			args: args{"simple"},
			want: "simple",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertHasuraTableName(tt.args.name); got != tt.want {
				t.Errorf("convertHasuraTableName() = %v, want %v", got, tt.want)
			}
		})
	}
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

	h := NewHasuraEngine(url)

	err := h.trackTable(DefaultHasuraSource, DefaultHasuraSource, "table1")
	if err != nil {
		t.Errorf("trackTable() should success, err=%v", err)
	}

	err = h.trackTable(DefaultHasuraSource, DefaultHasuraSource, "table1")
	if err == nil {
		t.Errorf("trackTable() should raise err")
	}
}

func Test_callHasura_trackTable(t *testing.T) {
	url := "http://127.0.0.1:8080/"
	h := NewHasuraEngine(url)
	if err := h.trackTable(DefaultHasuraSource, DefaultHasuraSchema, "profiles"); err != nil {
		t.Log(err)
	}
}

func Test_callHasura_untrackTable(t *testing.T) {
	url := "http://127.0.0.1:8080/"
	h := NewHasuraEngine(url)
	if err := h.untrackTable(DefaultHasuraSource, DefaultHasuraSchema, "profiles"); err != nil {
		t.Log(err)
	}
}

func Test_callHasura_addDefaultSource(t *testing.T) {
	url := "http://127.0.0.1:8082/"
	h := NewHasuraEngine(url)
	if err := h.addDefaultSourceAndSchema(); err != nil {
		t.Log(err)
	}
}

func Test_callHasura_addDefaultSchema(t *testing.T) {
	url := "http://127.0.0.1:8082/"
	h := NewHasuraEngine(url)
	if err := h.addDefaultSchema(); err != nil {
		t.Log(err)
	}
}
