package auth

import (
	"io"
	http2 "kwil/x/test/http"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuth_ServeHTTP(t *testing.T) {
	type fields struct {
		h http.Handler
		m Manager
	}
	type args struct {
		r *http.Request
	}

	km, _ := NewKeyManager(strings.NewReader(`{"keys": ["keya"]}`))
	testData := "dummy served"

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    string
	}{
		{
			name: "valid api key",
			fields: fields{
				h: &http2.DummyHttpHandler{Data: testData},
				m: km,
			},
			args: args{
				r: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					req.Header.Set(ApiKeyHeader, "keya")
					return req
				}(),
			},
			wantErr: false,
			want:    testData,
		},
		{
			name: "nonexist api key",
			fields: fields{
				h: &http2.DummyHttpHandler{Data: testData},
				m: km,
			},
			args: args{
				r: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					req.Header.Set(ApiKeyHeader, "keynotexist")
					return req
				}(),
			},
			wantErr: true,
			want:    MessageUnauthorized,
		},
		{
			name: "api key not present",
			fields: fields{
				h: &http2.DummyHttpHandler{Data: testData},
				m: km,
			},
			args: args{
				r: func() *http.Request {
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					return req
				}(),
			},
			wantErr: true,
			want:    MessageUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			a := MAuth(tt.fields.m)
			a.Middleware(tt.fields.h).ServeHTTP(w, tt.args.r)

			res := w.Result()
			defer res.Body.Close()

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if !tt.wantErr && string(data) != tt.want {
				t.Errorf("expected '%v' got '%v'", tt.want, string(data))
			}

			if tt.wantErr && res.StatusCode != http.StatusUnauthorized && tt.want != string(data) {
				t.Errorf("expected error got nil")
			}

		})
	}
}
