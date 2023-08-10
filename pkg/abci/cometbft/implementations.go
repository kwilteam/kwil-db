package cometbft

import (
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/privval"
)

// this file contains some basic implementations of cometbft validator interfaces.
// Some of these are actually hard to implement (like PrivKey), because some basic digging
// reveals that the internals of CometBFT might be tied to their own implementations
// I am just including these here to organize my thoughts and requirements

// this is just a placeholder
type CometBftPrivateKey struct {
}

func (c *CometBftPrivateKey) Bytes() []byte {
	panic("TODO")
}

func (c *CometBftPrivateKey) Equals(p0 crypto.PrivKey) bool {
	panic("TODO")
}

func (c *CometBftPrivateKey) PubKey() crypto.PubKey {
	panic("TODO")
}

func (c *CometBftPrivateKey) Sign(msg []byte) ([]byte, error) {
	panic("TODO")
}

// this Type seems to be the big hold up, since it essentially couples their internal implementations
func (c *CometBftPrivateKey) Type() string {
	panic("TODO")
}

// newPrivateValidator creates a new private validator with the given private key
// we don't need the filepaths, they are only used for the Save() method which is only used
// in testing and cometBFTs Cobra Commands.  Save() is not included in the interface required
// by NewNode()
func newPrivateValidator(pk *CometBftPrivateKey) *privval.FilePV {
	return privval.NewFilePV(pk, "", "") // save is not called, so we don't need to worry about the file paths
}
