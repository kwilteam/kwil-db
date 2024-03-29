package common

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

	err = SaveCookie(authFile, []byte("0x123"), &ck)
	assert.NoError(t, err)

	got, err := LoadPersistedCookie(authFile, []byte("0x123"))
	assert.NoError(t, err)

	assert.EqualValues(t, &ck, got)
}
