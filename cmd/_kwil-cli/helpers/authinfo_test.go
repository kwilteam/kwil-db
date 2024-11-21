package helpers

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadKGWAuthInfo_without_domain(t *testing.T) {
	// this test just to show what the old behavior was

	ckA := http.Cookie{
		Name:    "AAA",
		Value:   "AAA",
		Path:    "AAA",
		Domain:  "AAA",
		Expires: time.Date(2023, 10, 27, 15, 46, 58, 651387237, time.UTC),
	}

	ckB := http.Cookie{
		Name:    "BBB",
		Value:   "BBB",
		Path:    "BBB",
		Domain:  "BBB",
		Expires: time.Date(2023, 10, 27, 15, 46, 58, 651387237, time.UTC),
	}

	var err error
	authFile := filepath.Join(t.TempDir(), "auth.json")
	domain := ""

	// authn on site A
	err = SaveCookie(authFile, domain, []byte("0x123"), &ckA)
	assert.NoError(t, err)

	// authn on site B
	err = SaveCookie(authFile, domain, []byte("0x123"), &ckB)
	assert.NoError(t, err)

	got, err := LoadPersistedCookie(authFile, domain, []byte("0x123"))
	assert.NoError(t, err)

	// ckA has been overwritten by ckB
	assert.NotEqualValues(t, &ckA, got)
	assert.EqualValues(t, &ckB, got)
}

func TestLoadKGWAuthInfo(t *testing.T) {
	ck := http.Cookie{
		Name:       "test",
		Value:      "test",
		Path:       "test",
		Domain:     "test",
		Expires:    time.Date(2023, 10, 27, 15, 46, 58, 651387237, time.UTC),
		RawExpires: "",
		MaxAge:     0,
		Secure:     false,
		HttpOnly:   false,
		SameSite:   0,
		Raw:        "",
		Unparsed:   nil,
	}

	var err error
	authFile := filepath.Join(t.TempDir(), "auth.json")
	domain := "https://kgw.kwil.com"

	err = SaveCookie(authFile, domain, []byte("0x123"), &ck)
	assert.NoError(t, err)

	got, err := LoadPersistedCookie(authFile, domain, []byte("0x123"))
	assert.NoError(t, err)

	assert.EqualValues(t, &ck, got)
}

func Test_getDomain(t *testing.T) {
	type args struct {
		target string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantDoamin string
	}{
		// TODO: Add test cases.
		{
			name: "empty string",
			args: args{
				target: "",
			},
			wantErr:    true,
			wantDoamin: "",
		},
		{
			name: "localhost with port",
			args: args{
				target: "http://localhost:8080/api",
			},
			wantDoamin: "http://localhost:8080",
		},
		{
			name: "https localhost with port",
			args: args{
				target: "https://localhost:8080/api/",
			},
			wantDoamin: "https://localhost:8080",
		},
		{
			name: "http example.com",
			args: args{
				target: "http://example.com/a/b/c",
			},
			wantDoamin: "http://example.com",
		},
		{
			name: "https example.com",
			args: args{
				target: "https://example.com/a/b/c",
			},
			wantDoamin: "https://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, err := getDomain(tt.args.target)
			if tt.wantErr {
				assert.Errorf(t, err, "getDomain(%v)", tt.args.target)
			}
			assert.Equalf(t, tt.wantDoamin, domain, "getDomain(%v)", tt.args.target)
		})
	}
}
