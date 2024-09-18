package server

import (
	"testing"
)

func Test_cleanListenAddr(t *testing.T) {
	type args struct {
		addr        string
		defaultPort string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"ok no change",
			args{
				addr:        "tcp://127.0.0.1:9090",
				defaultPort: "8080",
			},
			"tcp://127.0.0.1:9090",
		},
		{
			"no port or scheme",
			args{
				addr:        "127.0.0.1",
				defaultPort: "8080",
			},
			"tcp://127.0.0.1:8080",
		},
		{
			"no scheme",
			args{
				addr:        "127.0.0.1:9090",
				defaultPort: "8080",
			},
			"tcp://127.0.0.1:9090",
		},
		{
			"ok no change",
			args{
				addr:        "tcp://localhost:9090",
				defaultPort: "8080",
			},
			"tcp://localhost:9090",
		},
		{
			"no port or scheme",
			args{
				addr:        "localhost",
				defaultPort: "8080",
			},
			"tcp://localhost:8080",
		},
		{
			"no scheme",
			args{
				addr:        "localhost:9090",
				defaultPort: "8080",
			},
			"tcp://localhost:9090",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanListenAddr(tt.args.addr, tt.args.defaultPort); got != tt.want {
				t.Errorf("cleanListenAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}
