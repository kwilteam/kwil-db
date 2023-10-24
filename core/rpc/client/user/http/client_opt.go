package http

import (
	"net/http"
)

type ClientOption interface {
	apply(*clientConfig)
}

type fnClientOption func(*clientConfig)

func (fn fnClientOption) apply(opt *clientConfig) {
	fn(opt)
}

func EnableDebugMode() ClientOption {
	return fnClientOption(func(cfg *clientConfig) {
		cfg.debugMode = true
	})
}

func WithHttpClient(httpClient *http.Client) ClientOption {
	return fnClientOption(func(cfg *clientConfig) {
		cfg.httpClient = httpClient
	})
}

func WithHeader(key, value string) ClientOption {
	return fnClientOption(func(cfg *clientConfig) {
		if cfg.httpHeaders == nil {
			cfg.httpHeaders = make(http.Header)
		}
		cfg.httpHeaders.Set(key, value)
	})
}

func WithCookie(cookie *http.Cookie) ClientOption {
	return fnClientOption(func(cfg *clientConfig) {
		if cfg.cookies == nil {
			cfg.cookies = make([]*http.Cookie, 0)
		}
		cfg.cookies = append(cfg.cookies, cookie)

	})
}

// clientConfig is the configuration for a Client.
type clientConfig struct {
	httpClient  *http.Client
	debugMode   bool
	httpHeaders http.Header
	cookies     []*http.Cookie
}
