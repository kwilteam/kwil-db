package adminclient

import (
	"github.com/kwilteam/kwil-db/core/log"
)

type AdminClientOpt func(*AdminClient)

// WithLogger sets the logger for the admin client.
func WithLogger(logger log.Logger) AdminClientOpt {
	return func(c *AdminClient) {
		c.log = logger
	}
}

// WithTLS provides the required TLS files for the admin client to connect via gRPC.
func WithTLS(kwildCertFile, clientKeyFile, clientCertFile string) AdminClientOpt {
	return func(c *AdminClient) {
		c.kwildCertFile = kwildCertFile
		c.clientKeyFile = clientKeyFile
		c.clientCertFile = clientCertFile
	}
}
