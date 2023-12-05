package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/internal/utils"
)

const (
	kgwAuthTokenFileName = "auth.json"
)

// KGWAuthTokenFilePath returns the path to the file that stores the Gateway Auth cookies.
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

func convertToCookie(c *http.Cookie) cookie {
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

func convertToHttpCookie(c cookie) *http.Cookie {
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

// PersistedCookies is a set of Gateway Auth cookies that can be saved to a file.
// It maps a base64 user identifier to a cookie, ensuring only one cookie per wallet.
// It uses a custom cookie type that is json serializable.
type PersistedCookies map[string]cookie

// LoadPersistedCookie loads a persisted cookie from the auth file.
// It will look up the cookie for the given user identifier.
// If nothing is found, it returns nil, nil.
func LoadPersistedCookie(authFile string, userIdentifier []byte) (*http.Cookie, error) {
	if !utils.FileExists(authFile) {
		return nil, nil
	}

	file, err := utils.CreateOrOpenFile(authFile)
	if err != nil {
		return nil, fmt.Errorf("open kgw auth file: %w", err)
	}

	var aInfo PersistedCookies
	err = json.NewDecoder(file).Decode(&aInfo)
	if err != nil {
		return nil, fmt.Errorf("unmarshal kgw auth file: %w", err)
	}

	b64Identifier := base64.StdEncoding.EncodeToString(userIdentifier)
	cookie := aInfo[b64Identifier]

	return convertToHttpCookie(cookie), nil
}

// SaveCookie saves the cookie to auth file.
// It will overwrite the cookie if the address already exists.
func SaveCookie(authFile string, userIdentifier []byte, originCookie *http.Cookie) error {
	cookie := convertToCookie(originCookie)

	authInfoBytes, err := utils.ReadOrCreateFile(authFile)
	if err != nil {
		return fmt.Errorf("read kgw auth file: %w", err)
	}

	b64Identifier := base64.StdEncoding.EncodeToString(userIdentifier)

	var aInfo PersistedCookies
	if len(authInfoBytes) == 0 {
		aInfo = make(PersistedCookies)
	} else {
		err = json.Unmarshal(authInfoBytes, &aInfo)
		if err != nil {
			return fmt.Errorf("unmarshal kgw auth file: %w", err)
		}
	}
	aInfo[b64Identifier] = cookie

	jsonBytes, err := json.MarshalIndent(&aInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal kgw auth info: %w", err)
	}

	err = utils.WriteFile(authFile, jsonBytes)
	if err != nil {
		return fmt.Errorf("write kgw auth file: %w", err)
	}
	return nil
}

// DeleteCookie will delete a cookie that exists for a given user identifier.
// If no cookie exists for the user identifier, it will do nothing.
func DeleteCookie(authFile string, userIdentifier []byte) error {
	authInfoBytes, err := utils.ReadOrCreateFile(authFile)
	if err != nil {
		return fmt.Errorf("read kgw auth file: %w", err)
	}

	b64Identifier := base64.StdEncoding.EncodeToString(userIdentifier)

	var aInfo PersistedCookies
	if len(authInfoBytes) == 0 {
		aInfo = make(PersistedCookies)
	} else {
		err = json.Unmarshal(authInfoBytes, &aInfo)
		if err != nil {
			return fmt.Errorf("unmarshal kgw auth file: %w", err)
		}
	}
	delete(aInfo, b64Identifier)

	jsonBytes, err := json.MarshalIndent(&aInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal kgw auth info: %w", err)
	}

	err = utils.WriteFile(authFile, jsonBytes)
	if err != nil {
		return fmt.Errorf("write kgw auth file: %w", err)
	}
	return nil
}
