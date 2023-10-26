package common

import (
	"encoding/json"
	"fmt"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/internal/utils"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const (
	kgwAuthTokenFileName = "kgw_auth.json"
)

func KGWAuthTokenFilePath() string {
	return filepath.Join(config.DefaultConfigDir, kgwAuthTokenFileName)
}

// cookie is a copy of http.Cookie struct, with explicit json tags
type cookie struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`

	Path       string    `json:"path,omitempty"`        // optional
	Domain     string    `json:"domain,omitempty"`      // optional
	Expires    time.Time `json:"expires"`               // optional
	RawExpires string    `json:"raw_expires,omitempty"` // for reading cookies only

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int      `json:"max_age,omitempty"`
	Secure   bool     `json:"secure,omitempty"`
	HttpOnly bool     `json:"http_only,omitempty"`
	SameSite int      `json:"same_site,omitempty"`
	Raw      string   `json:"raw,omitempty"`
	Unparsed []string `json:"unparsed,omitempty"` // Raw text of unparsed attribute-value pairs
}

func ConvertToCookie(c *http.Cookie) cookie {
	return cookie{
		Name:       c.Name,
		Value:      c.Value,
		Path:       c.Path,
		Domain:     c.Domain,
		Expires:    c.Expires,
		RawExpires: c.RawExpires,
		MaxAge:     c.MaxAge,
		Secure:     c.Secure,
		HttpOnly:   c.HttpOnly,
		SameSite:   int(c.SameSite),
		Raw:        c.Raw,
		Unparsed:   c.Unparsed,
	}
}

func ConvertToHttpCookie(c cookie) *http.Cookie {
	return &http.Cookie{
		Name:       c.Name,
		Value:      c.Value,
		Path:       c.Path,
		Domain:     c.Domain,
		Expires:    c.Expires,
		RawExpires: c.RawExpires,
		MaxAge:     c.MaxAge,
		Secure:     c.Secure,
		HttpOnly:   c.HttpOnly,
		SameSite:   http.SameSite(c.SameSite),
		Raw:        c.Raw,
		Unparsed:   c.Unparsed,
	}
}

// KGWAuthInfo represents the KGW authentication info for a wallet address
type KGWAuthInfo struct {
	Address string `json:"address"`
	Cookie  cookie `json:"cookie"`
}

// LoadKGWAuthInfo loads the KGW authentication info for a wallet address.
// If the address is not authenticated(local), it returns nil.
func LoadKGWAuthInfo(address string) (*KGWAuthInfo, error) {
	address = strings.ToLower(address)
	if !utils.FileExists(KGWAuthTokenFilePath()) {
		return nil, nil
	}

	authFile, err := utils.CreateOrOpenFile(KGWAuthTokenFilePath())
	if err != nil {
		return nil, fmt.Errorf("open kgw auth file: %w", err)
	}

	var aInfo []KGWAuthInfo
	err = json.NewDecoder(authFile).Decode(aInfo)
	if err != nil {
		return nil, fmt.Errorf("unmarshal kgw auth file: %w", err)
	}

	// always overwrite if the address already exists
	for _, a := range aInfo {
		if a.Address == address {
			return &a, nil
		}
	}

	return nil, nil
}
