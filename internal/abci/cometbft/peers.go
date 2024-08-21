package cometbft

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/internal/sql/versioning"
)

const (
	p2pSchemaName = "kwild_peers"

	p2pStoreVersion = 0

	initPeersTable = `CREATE TABLE IF NOT EXISTS ` + p2pSchemaName + `.peers (
		peer_id TEXT PRIMARY KEY
	);`

	addPeer = `INSERT INTO ` + p2pSchemaName + `.peers ` + `VALUES ($1);`

	removePeer = `DELETE FROM ` + p2pSchemaName + `.peers ` + `WHERE peer_id = $1;`

	listPeers = `SELECT peer_id FROM ` + p2pSchemaName + `.peers;`
)

var (
	ErrPeerAlreadyWhitelisted = fmt.Errorf("peer already whitelisted")
	ErrPeerNotWhitelisted     = fmt.Errorf("peer not whitelisted")
)

// PeerWhitelist object is used to manage the set of peers that a node is allowed to connect to.
// Whitelist peers are the list of allowed nodes that the node can accept connections from.
// Whitelist peers can be configured using the "chain.p2p.whitelist_peers" config in the config.toml file
// and can be updated dynamically using kwil-admin commands.
// The genesis validators are by default added to the whitelist peers during genesis. Any new validators
// are automatically added to the whitelist peers and demoted validators are removed from the whitelist.
// The Peers gossiped using PEX are not automatically trusted, and to be added manually if needed.
type PeerWhiteList struct {
	// privateMode is a flag to run the node in private mode. If disabled, the node will accept connections from any peer.
	// If enabled, the node will only accept connections from the current validators and whitelist peers.
	privateMode bool

	// whitelistPeers is a map of node IDs that the node can accept connections from.
	whitelistPeers map[string]bool
	peerMtx        sync.RWMutex // protects peers

	// removePeerFn is a function to gracefully stop and remove a peer connection.
	removePeerFn func(peerID string) error

	db sql.TxMaker
}

func initTables(ctx context.Context, tx sql.DB) error {
	_, err := tx.Execute(ctx, initPeersTable)
	return err
}

// NewPeers creates a new Peers object.
func P2PInit(ctx context.Context, db sql.TxMaker, privateMode bool, whitelistPeers []string) (*PeerWhiteList, error) {
	upgradeFns := map[int64]versioning.UpgradeFunc{
		0: initTables,
	}

	if err := versioning.Upgrade(ctx, db, p2pSchemaName, upgradeFns, p2pStoreVersion); err != nil {
		return nil, err
	}

	p := &PeerWhiteList{
		whitelistPeers: make(map[string]bool),
		db:             db,
		privateMode:    privateMode,
	}

	// Load peers from disk
	if err := p.loadPeers(ctx); err != nil {
		return nil, err
	}

	// add and persist peers from the whitelistPeers config if not already present
	if err := p.AddPeers(ctx, whitelistPeers); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *PeerWhiteList) AddPeers(ctx context.Context, peers []string) error {
	p.peerMtx.Lock()
	defer p.peerMtx.Unlock()

	tx, err := p.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, peer := range peers {
		peer = strings.ToLower(peer)
		_, ok := p.whitelistPeers[peer]
		if ok {
			continue
		}

		p.whitelistPeers[peer] = true
		// Persist peers to disk
		_, err = tx.Execute(ctx, addPeer, peer)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// AddPeer adds a peer to the whitelistPeers list. If the peer is already in the list,
// it returns an error indicating that the peer is already whitelisted.
func (p *PeerWhiteList) AddPeer(ctx context.Context, peer string) error {
	p.peerMtx.Lock()
	defer p.peerMtx.Unlock()

	peer = strings.ToLower(peer)
	_, ok := p.whitelistPeers[peer]
	if ok {
		return ErrPeerAlreadyWhitelisted
	}

	tx, err := p.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Persist peers to disk
	_, err = tx.Execute(ctx, addPeer, peer)
	if err != nil {
		return err
	}
	p.whitelistPeers[peer] = true

	return tx.Commit(ctx)
}

// RemovePeer gracefully stops an existing peer connection and removes the peer from the whitelisted peers list.
// If the peer is not found in the whitelistPeers list, it returns an error indicating that the peer is not found.
func (p *PeerWhiteList) RemovePeer(ctx context.Context, peer string) error {
	p.peerMtx.Lock()
	defer p.peerMtx.Unlock()

	peer = strings.ToLower(peer)
	_, ok := p.whitelistPeers[peer] // check if peer exists
	if !ok {
		return ErrPeerNotWhitelisted
	}

	tx, err := p.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Remove peer from the whitelisted peers list and from the database
	_, err = tx.Execute(ctx, removePeer, peer)
	if err != nil {
		return err
	}
	delete(p.whitelistPeers, peer)

	// Call removePeerFn to gracefully stop and remove the peer connection
	// if the node is already connected to the peer.
	if p.removePeerFn != nil {
		if err := p.removePeerFn(peer); err != nil {
			return fmt.Errorf("failed to remove peer %s: %w", peer, err)
		}
	}
	return tx.Commit(ctx)
}

// IsPeerWhitelisted checks if a peer is in the list of whitelisted peers.
// If the node is running with private mode disabled, it always returns true.
func (p *PeerWhiteList) IsPeerWhitelisted(peer string) bool {
	p.peerMtx.RLock()
	defer p.peerMtx.RUnlock()

	if !p.privateMode {
		// In Public mode, accept connections from any peer
		return true
	}

	peer = strings.ToLower(peer)
	// Check if peer is in the whitelistPeers
	_, ok := p.whitelistPeers[peer]
	return ok
}

// ListPeers returns the list of whitelisted peers.
func (p *PeerWhiteList) ListPeers(ctx context.Context) []string {
	p.peerMtx.RLock()
	defer p.peerMtx.RUnlock()

	var peers []string
	for peer := range p.whitelistPeers {
		peers = append(peers, peer)
	}

	return peers
}

// loadPeers loads the peers from the database into the whitelistPeers map.
func (p *PeerWhiteList) loadPeers(ctx context.Context) error {
	tx, err := p.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Execute(ctx, listPeers)
	if err != nil {
		return err
	}

	for _, row := range res.Rows {
		peer, ok := row[0].(string)
		if !ok {
			return fmt.Errorf("expected string for peer_id, got %T", row[0])
		}

		p.whitelistPeers[peer] = true
	}

	return tx.Commit(ctx)
}

func (p *PeerWhiteList) SetRemovePeerFn(fn func(peerID string) error) {
	p.removePeerFn = fn
}

// NodeIDAddressString makes a full CometBFT node ID address string in the
// format <nodeID>@hostPort where nodeID is derived from the provided public
// key.
func NodeIDAddressString(pubkey ed25519.PubKey, hostPort string) string {
	nodeID := p2p.PubKeyToID(pubkey)
	return p2p.IDAddressString(nodeID, hostPort)
}
