// package adminclient provides a client for the Kwil admin service.
// The admin service is used to perform node administrative actions,
// such as submitting validator transactions, retrieving node status, etc.
package adminclient

import (
	"testing"
)

func Test_prepareHTTPDialerURL(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantTarget string
		wantDialer bool
		wantErr    bool
	}{
		{
			"just ip",
			"127.0.0.1",
			"http://127.0.0.1:8485",
			false,
			false,
		},
		{
			"just hostname",
			"localhost",
			"http://localhost:8485",
			false,
			false,
		},
		{
			"http hostname",
			"http://localhost",
			"http://localhost:8485",
			false,
			false,
		},
		{
			"just ip:port",
			"127.0.0.1:8485",
			"http://127.0.0.1:8485",
			false,
			false,
		},
		{
			"https ip:port",
			"https://127.0.0.1:8485",
			"https://127.0.0.1:8485",
			false,
			false,
		},
		{
			"implicit unix",
			"/var/run/kwil.socket",
			"http://local.socket", // dialer captures the "host"
			true,
			false,
		},
		{
			"explicit unix",
			"unix:///var/run/kwil.socket",
			"http://local.socket",
			true,
			false,
		},
		{
			"http unix",
			"http:///var/run/kwil.socket",
			"http://local.socket", // dialer captures the "host"
			true,
			false,
		},
		{
			"https unix not supported",
			"https:///var/run/kwil.socket",
			"",
			false,
			true,
		},
		{
			"bad scheme",
			"asdf://badhost/bad",
			"",
			false,
			true, // error
		},
		{
			"pseudo-scheme known by url.ParseURL but unsupported by http client",
			"tcp://badhost/bad",
			"",
			false,
			true, // error
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, dialer, err := prepareHTTPDialerURL(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("prepareURLDialer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantTarget == "" {
				if target != nil && target.String() != "" {
					t.Errorf("prepareURLDialer() got = %v, want %v", target, tt.wantTarget)
				}
			} else if target.String() != tt.wantTarget {
				t.Errorf("prepareURLDialer() got = %v, want %v", target, tt.wantTarget)
			}
			if haveDialer := (dialer != nil); tt.wantDialer != haveDialer {
				t.Errorf("dialer incorrect (want %v, got %v)", tt.wantDialer, haveDialer)
			}
		})
	}
}
