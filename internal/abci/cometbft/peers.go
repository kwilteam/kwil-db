package cometbft

import (
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
)

// NodeIDAddressString makes a full CometBFT node ID address string in the
// format <nodeID>@hostPort where nodeID is derived from the provided public
// key.
func NodeIDAddressString(pubkey ed25519.PubKey, hostPort string) string {
	nodeID := p2p.PubKeyToID(pubkey)
	return p2p.IDAddressString(nodeID, hostPort)
}
