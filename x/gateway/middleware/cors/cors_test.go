package cors

import (
	http2 "kwil/x/test/http"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCors_ServeHTTP(t *testing.T) {
	type fields struct {
		h http.Handler
	}
	type args struct {
		r    *http.Request
		cors string
	}

	testData := "dummy served"
	testOrigin := "http://bar.example"
	testOrigins := "http://foo.example,http://bar.example"
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{
			name: "cors allow *",
			fields: fields{
				h: &http2.DummyHttpHandler{Data: testData},
			},
			args: args{
				r: func() *http.Request {
					req := httptest.NewRequest(http.MethodOptions, "/", nil)
					req.Header.Set("Origin", testOrigin)
					return req
				}(),
				cors: "*",
			},
			want: map[string]string{
				"Access-Control-Allow-Origin":  testOrigin,
				"Access-Control-Allow-Methods": AllowMethods,
				"Access-Control-Allow-Headers": AllowHeaders,
			},
		},
		{
			name: "cors allow origins return request origin",
			fields: fields{
				h: &http2.DummyHttpHandler{Data: testData},
			},
			args: args{
				r: func() *http.Request {
					req := httptest.NewRequest(http.MethodOptions, "/", nil)
					req.Header.Set("Origin", testOrigin)
					return req
				}(),
				cors: testOrigins,
			},
			want: map[string]string{
				"Access-Control-Allow-Origin":  testOrigin,
				"Access-Control-Allow-Methods": AllowMethods,
				"Access-Control-Allow-Headers": AllowHeaders,
			},
		},
		{
			name: "empty",
			fields: fields{
				h: &http2.DummyHttpHandler{Data: testData},
			},
			args: args{
				r: func() *http.Request {
					req := httptest.NewRequest(http.MethodOptions, "/", nil)
					req.Header.Set("Origin", testOrigin)
					return req
				}(),
				cors: "",
			},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c := newCors(tt.fields.h, tt.args.cors)
			c.ServeHTTP(w, tt.args.r)

			res := w.Result()
			defer res.Body.Close()

			for k, v := range tt.want {
				if res.Header.Get(k) != v {
					t.Errorf("expect header '%s=%s', got '%s'", k, v, res.Header.Get(k))
				}
			}
		})
	}
}

func Test_allowedOrigin(t *testing.T) {
	type args struct {
		cors   string
		origin string
	}
	testOrigin := "http://bar.example"
	testOrigins := "http://foo.example,http://bar.example"
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "*",
			args: args{
				cors:   "*",
				origin: testOrigin,
			},
			want: true,
		},
		{
			name: "allowed with one origin",
			args: args{
				cors:   testOrigin,
				origin: testOrigin,
			},
			want: true,
		},
		{
			name: "allowed with multi origins",
			args: args{
				cors:   testOrigins,
				origin: testOrigin,
			},
			want: true,
		},
		{
			name: "not allowed with one origin",
			args: args{
				cors:   testOrigin,
				origin: "http://baz.example",
			},
			want: false,
		},
		{
			name: "not allowed with multi origins",
			args: args{
				cors:   testOrigins,
				origin: "http://baz.example",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := allowedOrigin(tt.args.cors, tt.args.origin); got != tt.want {
				t.Errorf("allowedOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}
