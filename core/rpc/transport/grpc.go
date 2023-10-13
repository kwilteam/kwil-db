package transport

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// @yaiba TODO: make this configurable
	// DefaultDialTimeout is the default dial timeout.
	DefaultDialTimeout = 3 * time.Second
	// DefaultRequestTimeout is the default request timeout.
	DefaultRequestTimeout = 3 * time.Second
)

type GrpcTransporter interface {
	// Dial connects to the given address.
	Dial(ctx context.Context, address string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

type TlsOption struct{}

type TimeOutOption struct {
	Dial    time.Duration
	Request time.Duration
}

type GrpcTransport struct {
	tlsOpts *TlsOption
	timeout TimeOutOption
}

func NewGrpcTransport(tlsOpts *TlsOption, timeout TimeOutOption) *GrpcTransport {
	return &GrpcTransport{
		tlsOpts: tlsOpts,
		timeout: timeout,
	}
}

func (t *GrpcTransport) Dial(ctx context.Context, address string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	options := append([]grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
			d := &net.Dialer{}
			conn, err := d.DialContext(ctx, "tcp", address)
			if err != nil {
				return nil, err
			}
			return &reqTimeoutConn{
				conn:    conn,
				timeout: t.timeout.Request,
			}, nil
		})}, opts...)

	ctx, cancel := context.WithTimeout(ctx, t.timeout.Dial)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, options...)
	if err == context.Canceled {
		return nil, err
	}
	if err != nil && strings.Contains(err.Error(), "connection refused") {
		return nil, fmt.Errorf("connection refused: %s", address)
	}

	return conn, err
}

func Dial(ctx context.Context, address string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	timeout := TimeOutOption{
		Dial:    DefaultDialTimeout,
		Request: DefaultRequestTimeout,
	}
	return NewGrpcTransport(nil, timeout).Dial(ctx, address, opts...)
}
