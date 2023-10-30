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

type LocalKGWAuthInfo []*KGWAuthInfo

// LoadKGWAuthInfo loads the address KGW authentication info from the given file.
// If the address is not authenticated(local), it returns nil.
func LoadKGWAuthInfo(authFile string, address string) (*KGWAuthInfo, error) {
	address = strings.ToLower(address)
	if !utils.FileExists(authFile) {
		return nil, nil
	}

	file, err := utils.CreateOrOpenFile(authFile)
	if err != nil {
		return nil, fmt.Errorf("open kgw auth file: %w", err)
	}

	var aInfo LocalKGWAuthInfo
	err = json.NewDecoder(file).Decode(&aInfo)
	if err != nil {
		return nil, fmt.Errorf("unmarshal kgw auth file: %w", err)
	}

	// always overwrite if the address already exists
	for _, a := range aInfo {
		if a.Address == address {
			return a, nil
		}
	}

	return nil, nil
}

// SaveAuthInfo saves the cookie to auth file.
func SaveAuthInfo(authFile string, address string, originCookie *http.Cookie) error {
	address = strings.ToLower(address)
	cookie := ConvertToCookie(originCookie)

	authInfoBytes, err := utils.ReadOrCreateFile(authFile)
	if err != nil {
		return fmt.Errorf("read kgw auth file: %w", err)
	}

	var aInfo LocalKGWAuthInfo

	if len(authInfoBytes) != 0 {
		// if the file is not empty, load the content
		err = json.Unmarshal(authInfoBytes, &aInfo)
		if err != nil {
			return fmt.Errorf("unmarshal kgw auth file: %w", err)
		}

		exists := false
		// always overwrite if the address already exists
		for _, a := range aInfo {
			if a.Address == address {
				a.Cookie = cookie
				exists = true
				break
			}
		}

		if !exists {
			aInfo = append(aInfo, &KGWAuthInfo{
				Address: address,
				Cookie:  cookie,
			})
		}
	} else {
		aInfo = append(aInfo, &KGWAuthInfo{
			Address: address,
			Cookie:  cookie,
		})
	}

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
