package client

import "net/http"

// ActionCallOpts is the options for action call.
// Currently only HTTP RPCClient supports this option.
type ActionCallOpts struct {
	// authn cookie, for provider that supports cookie authn, now only for KGW
	// Call action(view action) is read oriented; for data privacy, a network needs
	// extra infra to protect the data. KGW is such infra using cookie authn.
	// NOTE: setting cookie this way means the cookie policy is not applied.
	// AuthCookies is a general way to use cookie in SDK when calling action.
	AuthCookies []*http.Cookie
}

type ActionCallOption func(*ActionCallOpts)

func WithAuthCookie(cookie *http.Cookie) ActionCallOption {
	return func(opts *ActionCallOpts) {
		if opts.AuthCookies == nil {
			opts.AuthCookies = make([]*http.Cookie, 0)
		}
		opts.AuthCookies = append(opts.AuthCookies, cookie)
	}
}
