// package adminclient provides a client for the Kwil admin service.
// The admin service is used to perform node administrative actions,
// such as submitting validator transactions, retrieving node status, etc.
package adminclient

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"

	"github.com/kwilteam/kwil-db/core/log"
	admingrpc "github.com/kwilteam/kwil-db/core/rpc/client/admin/grpc"
	txGrpc "github.com/kwilteam/kwil-db/core/rpc/client/user/grpc"
	"github.com/kwilteam/kwil-db/core/rpc/transport"
	"github.com/kwilteam/kwil-db/core/types"
	adminTypes "github.com/kwilteam/kwil-db/core/types/admin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// AdminClient is a client for the Kwil admin service.
// It inherits both the admin and tx services.
type AdminClient struct {
	adminTransport // transport for admin client. we can just expose this, since we don't need to wrap it with any logic
	txClient       // should be the subset of the interface of core/client/Client that we want to expose here.

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
}

// adminTransport is a transport interface for the admin service.
type adminTransport interface {
	// Approve approves a node to join the network.
	// It returns a transaction hash.
	Approve(ctx context.Context, publicKey []byte) ([]byte, error)

	// Join submits a join request to the network, from the connected node.
	// It returns a transaction hash.
	Join(ctx context.Context) ([]byte, error)

	// JoinStatus returns the status of a join request.
	JoinStatus(ctx context.Context, publicKey []byte) (*types.JoinRequest, error)

	// Leave submits a leave request to the network, from the connected node.
	// It returns a transaction hash.
	Leave(ctx context.Context) ([]byte, error)

	// ListValidators returns the current validator set from the connected node.
	ListValidators(ctx context.Context) ([]*types.Validator, error)

	// Peers returns the current peer set from the connected node.
	Peers(ctx context.Context) ([]*adminTypes.PeerInfo, error)

	// Remove votes to remove a node from the network.
	Remove(ctx context.Context, publicKey []byte) ([]byte, error)

	// Status returns the current status of the connected node.
	Status(ctx context.Context) (*adminTypes.Status, error)

	// Version returns the current version of the connected node.
	Version(ctx context.Context) (string, error)
}

// New creates a new admin client.
// It can be configured to either use TLS or not, if using gRPC.
// The target arg should be either "tcp://localhost:50151", "localhost:50151", or "unix://path/to/socket.sock"
func New(ctx context.Context, target string, opts ...AdminClientOpt) (*AdminClient, error) {
	c := &AdminClient{
		log: log.NewNoOp(),
	}

	parsedTarget, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	// we can have:
	// tcp + tls
	// tcp + no tls
	// unix socket + no tls
	dialOpts := []grpc.DialOption{}

	switch parsedTarget.Scheme {
	case "tcp", "": // default to grpc
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
	case "unix":
		dialOpts = append(dialOpts, grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return net.Dial("unix", s)
		}))
	}

	// we dial a normal grpc connection, and then wrap it with the services
	conn, err := grpc.DialContext(ctx, target, dialOpts...)
	if err != nil {
		return nil, err
	}

	c.adminTransport = admingrpc.NewAdminClient(conn)

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
