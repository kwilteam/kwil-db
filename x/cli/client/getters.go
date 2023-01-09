package client

import (
	"crypto/ecdsa"
	"kwil/x/chain"
	chainClient "kwil/x/chain/client"
	"kwil/x/contracts/escrow"
	"kwil/x/contracts/token"
)

// unconnected client getters

func (c *unconnectedClient) Address() string {
	return c.address
}

func (c *unconnectedClient) ChainCode() chain.ChainCode {
	return c.chainCode
}

func (c *unconnectedClient) ChainClient() chainClient.ChainClient {
	return c.chainClient
}

func (c *unconnectedClient) Escrow() escrow.EscrowContract {
	return c.escrow
}

func (c *unconnectedClient) PoolAddress() string {
	return c.poolAddress
}

func (c *unconnectedClient) PrivateKey() *ecdsa.PrivateKey {
	return c.privateKey
}

func (c *unconnectedClient) Token() token.TokenContract {
	return c.token
}

func (c *unconnectedClient) TokenAddress() string {
	return c.tokenAddress
}

func (c *unconnectedClient) ValidatorAddress() string {
	return c.validatorAddress
}

// client getters
// this will just be calling the above getters

func (c *client) Address() string {
	return c.UnconnectedClient.Address()
}

func (c *client) ChainCode() chain.ChainCode {
	return c.UnconnectedClient.ChainCode()
}

func (c *client) ChainClient() chainClient.ChainClient {
	return c.UnconnectedClient.ChainClient()
}

func (c *client) Escrow() escrow.EscrowContract {
	return c.UnconnectedClient.Escrow()
}

func (c *client) PoolAddress() string {
	return c.UnconnectedClient.PoolAddress()
}

func (c *client) PrivateKey() *ecdsa.PrivateKey {
	return c.UnconnectedClient.PrivateKey()
}

func (c *client) Token() token.TokenContract {
	return c.UnconnectedClient.Token()
}

func (c *client) TokenAddress() string {
	return c.UnconnectedClient.TokenAddress()
}

func (c *client) ValidatorAddress() string {
	return c.UnconnectedClient.ValidatorAddress()
}
