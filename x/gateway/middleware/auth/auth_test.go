package auth

import (
	"net/http"
	"testing"
)

func TestAuth_ServeHTTP(t *testing.T) {
	type fields struct {
		h http.Handler
		m Manager
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Auth{
				h: tt.fields.h,
				m: tt.fields.m,
			}
			a.ServeHTTP(tt.args.w, tt.args.r)
		})
	}
}
