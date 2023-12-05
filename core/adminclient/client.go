// package adminclient provides a client for the Kwil admin service.
// The admin service is used to perform node administrative actions,
// such as submitting validator transactions, retrieving node status, etc.
package adminclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/kwilteam/kwil-db/core/log"
	adminRpc "github.com/kwilteam/kwil-db/core/rpc/client/admin"
	admingrpc "github.com/kwilteam/kwil-db/core/rpc/client/admin/grpc"
	txGrpc "github.com/kwilteam/kwil-db/core/rpc/client/user/grpc"
	"github.com/kwilteam/kwil-db/core/rpc/transport"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils/url"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// AdminClient is a client for the Kwil admin service.
// It inherits both the admin and tx services.
type AdminClient struct {
	adminRpc.AdminClient // transport for admin client. we can just expose this, since we don't need to wrap it with any logic
	txClient             // should be the subset of the interface of core/client/Client that we want to expose here.

	log log.Logger

	// tls cert files, if using grpc and not unix socket
	kwildCertFile  string
	clientKeyFile  string
	clientCertFile string
}

// txClient is the txsvc client interface.
// It allows us to selectively expose the txsvc client methods.
type txClient interface {
	// Ping pings the connected node.
	Ping(ctx context.Context) (string, error)
	// TxQuery queries a transaction by hash.
	TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error)
}

// NewClient creates a new admin client.
// It can be configured to either use TLS or not, if using gRPC.
// The target arg should be either "tcp://localhost:50151", "localhost:50151", or "unix://path/to/socket.sock"
func NewClient(ctx context.Context, target string, opts ...AdminClientOpt) (*AdminClient, error) {
	c := &AdminClient{
		log: log.NewNoOp(),
	}

	parsedTarget, err := url.ParseURL(target)
	if err != nil {
		return nil, err
	}

	// we can have:
	// tcp + tls
	// tcp + no tls
	// unix socket + no tls
	dialOpts := []grpc.DialOption{}

	switch parsedTarget.Scheme {
	case url.TCP: // default to grpc
		if c.kwildCertFile != "" || c.clientKeyFile != "" || c.clientCertFile != "" {
			// tcp + tls

			tlsCfg, err := newAuthenticatedTLSConfig(c.kwildCertFile, c.clientCertFile, c.clientKeyFile)
			if err != nil {
				return nil, err
			}

			dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
		} else {
			// tcp + no tls
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}
	case url.UNIX:
		dialOpts = append(dialOpts, grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return net.Dial("unix", s)
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	default:
		return nil, fmt.Errorf("unknown scheme %q", parsedTarget.Scheme)
	}

	// we dial a normal grpc connection, and then wrap it with the services
	conn, err := grpc.DialContext(ctx, parsedTarget.Target, dialOpts...)
	if err != nil {
		return nil, err
	}

	c.AdminClient = admingrpc.NewAdminClient(conn)

	c.txClient = txGrpc.WrapConn(conn)

	return c, nil
}

// newAuthenticatedTLSConfig creates a new tls.Config for an
// mutually-authenticated TLS (mTLS) client. In addition to the server's
// certificate file, the client's own key pair is required to support protocol
// level client authentication.
func newAuthenticatedTLSConfig(kwildCertFile, clientCertFile, clientKeyFile string) (*tls.Config, error) {
	cfg, err := transport.NewClientTLSConfigFromFile(kwildCertFile)
	if err != nil {
		return nil, err
	}

	authCert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return nil, err
	}
	cfg.Certificates = append(cfg.Certificates, authCert)

	return cfg, nil
}
