package http

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

func Test_cookie(t *testing.T) {
	jar, _ := cookiejar.New(nil)
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:   "test",
		Value:  "tv",
		Path:   "/",
		Domain: "example.com",
	}
	cookies = append(cookies, cookie)
	u, _ := url.Parse("http://example.com/search/")
	jar.SetCookies(u, cookies)
	fmt.Println(jar.Cookies(u))
}

func Test_cookie2(t *testing.T) {

	jar, _ := cookiejar.New(nil)
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:   "test",
		Value:  "tv",
		Path:   "/",
		Domain: "localhost:8080",
	}
	cookies = append(cookies, cookie)
	u, _ := url.Parse("localhost:8080/search/")
	jar.SetCookies(u, cookies)
	fmt.Println(jar.Cookies(u))
}

func Test_cookie3(t *testing.T) {

	jar, _ := cookiejar.New(nil)
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:   "test",
		Value:  "tv",
		Path:   "/",
		Domain: "localhost:8080",
	}
	cookies = append(cookies, cookie)
	u, _ := url.Parse("http://localhost:8080/search/")
	jar.SetCookies(u, cookies)
	fmt.Println(jar.Cookies(u))
}

func Test_cookie4(t *testing.T) {
	// we can test locally like this, set local dns point example.com to localhost
	jar, _ := cookiejar.New(nil)
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:   "test",
		Value:  "tv",
		Path:   "/",
		Domain: "example.com",
	}
	cookies = append(cookies, cookie)
	u, _ := url.Parse("http://example.com:8080/search/")
	jar.SetCookies(u, cookies)
	fmt.Println(jar.Cookies(u))
}
