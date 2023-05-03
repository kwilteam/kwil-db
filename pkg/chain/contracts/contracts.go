package contracts

import (
	"kwil/pkg/chain/contracts/escrow"
	"kwil/pkg/chain/contracts/token"
	"kwil/pkg/chain/provider"
)

type Contracter interface {
	Escrow(address string, opts ...escrow.EscrowOpts) (escrow.EscrowContract, error)
	Token(address string, opts ...token.TokenOpts) (token.TokenContract, error)
}

type contractBuilder struct {
	provider provider.ChainProvider
}

func New(provider provider.ChainProvider) Contracter {
	return &contractBuilder{
		provider: provider,
	}
}

func (c *contractBuilder) Escrow(address string, opts ...escrow.EscrowOpts) (escrow.EscrowContract, error) {
	return escrow.New(c.provider, address, opts...)
}

func (c *contractBuilder) Token(address string, opts ...token.TokenOpts) (token.TokenContract, error) {
	return token.New(c.provider, address, opts...)
}
