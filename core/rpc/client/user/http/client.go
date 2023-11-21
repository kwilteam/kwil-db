package http

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptrace"
	"net/url"
	"slices"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	contentType = "application/json"
)

// Client represent a connection to a kwil node
type Client struct {
	target    string // host:port or domain
	debugMode bool
	// https://cs.opensource.google/go/go/+/refs/tags/go1.21.3:src/net/http/client.go;l=562
	// https://cs.opensource.google/go/go/+/refs/tags/go1.21.3:src/net/http/response.go;l=59
	// for the connection to be reused, just close the resp.body is not enough,
	// the body also has to be read fully.
	conn    *http.Client
	headers http.Header
}

// Dial creates an HTTP connection to the given target.
// Supported URL schemes are http and https.
// For more configuration options, use DialOptions.
func Dial(target string) (*Client, error) {
	return DialOptions(target)
}

// DialOptions creates an HTTP connection to the given options.
func DialOptions(target string, opts ...ClientOption) (*Client, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target: %w", err)
	}

	switch u.Scheme {
	case "http", "https":
	default:
		return nil, fmt.Errorf("URL scheme not support: %s", u.Scheme)
	}

	cfg := new(clientConfig)
	for _, opt := range opts {
		opt.apply(cfg)
	}

	headers := make(http.Header, 2+len(cfg.httpHeaders))
	headers.Set("Accept", contentType)
	headers.Set("Content-Type", contentType)
	for key, values := range cfg.httpHeaders {
		headers[key] = values
	}

	clt := cfg.httpClient
	if clt == nil {
		clt = DefaultHTTPClient()
	}

	// enable cookie jar
	if clt.Jar == nil {
		clt.Jar = newCookieJar()
	}

	clt.Jar.SetCookies(u, cfg.cookies)

	client := &Client{
		target:    target,
		debugMode: cfg.debugMode,
		conn:      clt,
		headers:   headers,
	}

	return client, nil
}

// makeRequest makes a request to the target and returns the response body
// the caller should read all data in the body and close the response body
func (c *Client) makeRequest(req *http.Request) (*http.Response, error) {
	carriedCookies := req.Cookies()
	// default headers
	req.Header = c.headers.Clone()
	s := time.Now()

	var dnsStart, connStart, reqStart time.Time
	var dnsDuration, connDuration, reqDuration time.Duration
	var connReused bool
	if c.debugMode {
		trace := &httptrace.ClientTrace{
			DNSStart: func(_ httptrace.DNSStartInfo) {
				dnsStart = time.Now()
			},
			DNSDone: func(_ httptrace.DNSDoneInfo) {
				dnsDuration = time.Since(dnsStart)
			},
			GetConn: func(_ string) {
				connStart = time.Now()
			},
			GotConn: func(info httptrace.GotConnInfo) {
				if !info.Reused {
					connDuration = time.Since(connStart)
				} else {
					connReused = true
				}
				reqStart = time.Now()
			},
			WroteRequest: func(_ httptrace.WroteRequestInfo) {
				reqDuration = time.Since(reqStart)
			},
		}
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	}

	// handle jar cookies and request cookies
	if carriedCookies != nil {
		jarCookies := c.conn.Jar.Cookies(req.URL)
		if jarCookies == nil {
			// use request cookies if jar has no cookies
			for _, cookie := range carriedCookies {
				req.AddCookie(cookie)
			}
		} else {
			// NOTE: SetCookies won't overwrite existing cookies of cookieJar,
			// so we have to create a new cookie jar
			oldJar := c.conn.Jar
			defer func() {
				c.conn.Jar = oldJar
			}()

			// overwrite jar cookies with request cookies
			newJarCookies := append(carriedCookies, jarCookies...) // order matters
			// remove duplicated cookies; CompactFunc will modify the slice and
			// keep the first one of duplicated cookies, e.g., from carryCookies
			newJarCookies = slices.CompactFunc(newJarCookies,
				func(c1 *http.Cookie, c2 *http.Cookie) bool {
					return c1.Name == c2.Name
				})

			c.conn.Jar = newCookieJar()
			c.conn.Jar.SetCookies(req.URL, newJarCookies)
		}
	}

	resp, err := c.conn.Do(req)
	// NOTE: we can copy&close resp.Body to a buffer here so that later we
	// don't need to close the resp.Body to ensure the connection can be reused.
	// seems not a good idea

	if c.debugMode {
		duration := time.Since(s)
		// NOTE: kind of exotic using fmt here, but it's just for debugging
		fmt.Printf("request %s completed in %s (dns=%s, conn=%s, req=%s), reused=%t\n",
			req.URL.Path, duration, dnsDuration, connDuration, reqDuration, connReused)
	}

	return resp, err
}

func (c *Client) GetTarget() string {
	return c.target
}

func (c *Client) Close() error {
	c.conn = nil
	return nil
}

func DefaultHTTPClient() *http.Client {
	// same transport same connection
	tr := &http.Transport{
		//MaxConnsPerHost: 5, // default is 0, no limit
		//MaxIdleConnsPerHost: 2, // default is 2
		//DisableCompression: ,
		//DisableKeepAlives:  ,
		IdleConnTimeout: time.Second * 5, // default is 90s
	}
	return &http.Client{
		Transport: tr,
		Timeout:   time.Second * 3,
	}
}

func newCookieJar() http.CookieJar {
	// NOTE: not entirely sure about the PublicSuffixList
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	return jar
}
