// package url provides url fuctionalities to provide consistent
// parsing for Kwil clients.
package url

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// URL is a parsed URL.
type URL struct {
	// Original is the original URL string.
	Original string
	// Scheme is the protocol scheme, such as http or tcp.
	Scheme Scheme
	// Target is either the host, host:port, or unix socket path.
	Target string
	// Port is the port number, or 0 if not specified.
	Port int
}

// Scheme is a protocol scheme, such as http or tcp.
type Scheme string

func (s Scheme) valid() bool {
	switch s {
	case HTTP, HTTPS, TCP, UNIX:
		return true
	default:
		return false
	}
}

func (s Scheme) String() string {
	return string(s)
}

const (
	HTTP  Scheme = "http"
	HTTPS Scheme = "https"
	TCP   Scheme = "tcp"
	UNIX  Scheme = "unix"
)

// ParseURL parses a URL string into a URL struct.
// URLs can be of the form:
// - http://localhost:8080
// - tcp://localhost:8080
// - localhost:8080
// - localhost
// - unix:///var/run/kwil.sock
// If the URL does not have a scheme, it is assumed to be a tcp address.
func ParseURL(u string) (*URL, error) {
	original := u
	// if the url does not have a scheme, assume it's a tcp address
	hasScheme, err := hasScheme(u)
	if err != nil {
		return nil, err
	}
	if !hasScheme {
		u = "tcp://" + u
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	scheme := Scheme(parsed.Scheme)
	if !scheme.valid() { // I don't think this can error, but just in case
		return nil, fmt.Errorf("%w: %s", ErrUnknownScheme, scheme)
	}

	var target string
	switch scheme {
	case UNIX:
		target, err = expandPath(parsed.Path)
		if err != nil {
			return nil, err
		}
	default:
		target = parsed.Host
	}

	portString := parsed.Port()
	if portString == "" {
		portString = "0"
	}
	port, err := strconv.ParseInt(portString, 10, 32)
	if err != nil {
		return nil, err
	}

	return &URL{
		Original: original,
		Scheme:   scheme,
		Target:   target,
		Port:     int(port),
	}, nil
}

// hasScheme returns true if the url has a known scheme.
func hasScheme(u string) (bool, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return false, err
	}

	switch parsed.Scheme {
	case "tcp", "unix", "http", "https":
		return true, nil
	default:
		// see if it can be split by ://
		split := strings.Split(u, "://")
		if len(split) == 2 {
			return false, fmt.Errorf("%w: %s", ErrUnknownScheme, parsed.Scheme)
		}
		return false, nil
	}
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return strings.Replace(path, "~", home, 1), nil
	}
	return path, nil
}
