package cors

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	http2 "github.com/kwilteam/kwil-db/pkg/grpc/gateway/middleware/http-dummy"
)

func TestCors_ServeHTTP(t *testing.T) {
	type fields struct {
		h http.Handler
	}
	type args struct {
		r    *http.Request
		cors []string
	}

	testData := "dummy served"
	testOrigin := "http://bar.example"
	testOrigins := "http://foo.example,http://bar.example"
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantHeader map[string]string
		wantBody   string
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
				cors: []string{"*"},
			},
			wantHeader: map[string]string{
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
				cors: []string{testOrigins},
			},
			wantHeader: map[string]string{
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
				cors: []string{""},
			},
			wantHeader: map[string]string{},
		},
		{
			name: "non options method",
			fields: fields{
				h: &http2.DummyHttpHandler{Data: testData},
			},
			args: args{
				r: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					return req
				}(),
				cors: []string{""},
			},
			wantBody: testData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c := MCors(tt.args.cors)
			c.Middleware(tt.fields.h).ServeHTTP(w, tt.args.r)

			res := w.Result()
			defer res.Body.Close()

			if tt.wantHeader != nil {
				for k, v := range tt.wantHeader {
					if res.Header.Get(k) != v {
						t.Errorf("expect header '%s=%s', got '%s'", k, v, res.Header.Get(k))
					}
				}
			}

			if tt.wantBody != "" {
				data, err := io.ReadAll(res.Body)
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}

				if string(data) != tt.wantBody {
					t.Errorf("expected '%v' got '%v'", tt.wantBody, string(data))
				}
			}
		})
	}
}

func Test_allowedOrigin(t *testing.T) {
	type args struct {
		cors   []string
		origin string
	}
	testOrigin1 := "http://bar.example"
	testOrigin2 := "http://foo.example"
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "*",
			args: args{
				cors:   []string{"*"},
				origin: testOrigin1,
			},
			want: true,
		},
		{
			name: "allowed with one origin",
			args: args{
				cors:   []string{testOrigin1},
				origin: testOrigin1,
			},
			want: true,
		},
		{
			name: "allowed with multi origins",
			args: args{
				cors:   []string{testOrigin1, testOrigin2},
				origin: testOrigin1,
			},
			want: true,
		},
		{
			name: "not allowed with one origin",
			args: args{
				cors:   []string{testOrigin1},
				origin: "http://baz.example",
			},
			//want: false,
			want: true, // allow all

		},
		{
			name: "not allowed with multi origins",
			args: args{
				cors:   []string{testOrigin1, testOrigin2},
				origin: "http://baz.example",
			},
			//want: false,
			want: true, // allow all
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := allowedOrigin(tt.args.cors, tt.args.origin); got != tt.want {
				t.Errorf("allowedOrigin() = %v, wantHeader %v", got, tt.want)
			}
		})
	}
}
