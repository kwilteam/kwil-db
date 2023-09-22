package admin

import (
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/log"
)

type ClientOpt func(*Client)

func WithLogger(logger log.Logger) ClientOpt {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithSigner(signer crypto.Signer) ClientOpt {
	return func(c *Client) {
		c.signer = signer
	}
}
