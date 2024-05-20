package rpcserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrTo[T any](x T) *T {
	return &x
}

func Test_zeroID(t *testing.T) {
	var i any = (*int)(nil) // i != nil, it's a non-nil interface with nil data
	tests := []struct {
		name string
		id   any
		want bool
	}{
		{"int 0", int(0), true},
		{"int64 0", int64(0), true},
		{"float64 0", float64(0), true},
		{"ptr to int 0", ptrTo(0), true},
		{"nil ptr", (*int)(nil), true},
		{"non-interface to nil", i, true},
		{"nil", nil, true},
		{"empty string", "", true},
		{"int 1`", int(1), false},
		{"float64 1.1", float64(1.1), false},
		{"ptr to int 1", ptrTo(1), false},
		{"string a", "a", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := zeroID(tt.id); got != tt.want {
				t.Errorf("zeroID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_timeout(t *testing.T) {
	// This handler will simulate a request that exceeds the timeout.
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK) // if test passes, should not get this!
	})

	// Wrap that handler with a 500ms timeout.
	h = jsonRPCTimeoutHandler(h, 500*time.Millisecond, log.NewStdOut(log.DebugLevel))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(w, r)

	// Expect http.TimeoutHandler to have responded...
	assert.Equal(t, http.StatusServiceUnavailable, w.Result().StatusCode)

	// ...with our jsonrpc.Error.Code
	var resp jsonrpc.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, resp.Error.Code, jsonrpc.ErrorTimeout)
}
