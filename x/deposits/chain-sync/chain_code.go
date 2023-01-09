package chainsync

import (
	chains "kwil/x/chain"
)

func (c *chain) ChainCode() chains.ChainCode {
	return c.chainClient.ChainCode()
}
