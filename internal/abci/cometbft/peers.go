package cometbft

import (
	"context"
	"fmt"
	"sync"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/internal/sql/versioning"
)

const (
	p2pSchemaName = "peers" // TODO: Umm, should this be part of the chain metadata schema?

	p2pStoreVersion = 0

	initPeersTable = `CREATE TABLE IF NOT EXISTS ` + p2pSchemaName + `.peers (
		peer_id TEXT PRIMARY KEY
	);`

	addPeer = `INSERT INTO ` + p2pSchemaName + `.peers ` + `VALUES ($1) ` + `ON CONFLICT (peer_id) DO NOTHING;`

	removePeer = `DELETE FROM ` + p2pSchemaName + `.peers ` + `WHERE peer_id = $1;`
)

// Peers object is used to manage the set of peers that a node is allowed to connect to.
// Static peers are those that are added in the genesis and are trusted by default. These
// include the initial validators, persistent peers, and seed nodes.
// Whitelist peers are the list of allowed sentry nodes that a node can connect to.
// Whitelist peers can be updated dynamically using kwil-admin commands and are persisted
// to disk.
// The Peers gossiped using PEX are not automatically trusted, and to be added manually if needed.
// Any new validators are automatically added to the whitelist peers and demoted validators are removed from the whitelist.
type Peers struct {
	// whitelistPeers is a map of node IDs.
	whitelistPeers map[string]bool
	peerMtx        sync.RWMutex // protects peers

	removePeerFn func(peerID string) error

	db sql.TxMaker
}

func initTables(ctx context.Context, tx sql.DB) error {
	_, err := tx.Execute(ctx, initPeersTable)
	return err
}

// NewPeers creates a new Peers object.
func P2PInit(ctx context.Context, db sql.TxMaker, whitelistPeers []string) (*Peers, error) {
	upgradeFns := map[int64]versioning.UpgradeFunc{
		0: initTables,
	}

	if err := versioning.Upgrade(ctx, db, p2pSchemaName, upgradeFns, p2pStoreVersion); err != nil {
		return nil, err
	}

	p := &Peers{
		whitelistPeers: make(map[string]bool),
		db:             db,
	}

	// no need to persist them and these are added only once at startup
	for _, peer := range whitelistPeers {
		p.whitelistPeers[peer] = true
	}

	// Load peers from disk
	if err := p.loadPeers(ctx); err != nil {
		return nil, err
	}

	// load the whitelist peers from the config and persist if not already present
	for _, peer := range whitelistPeers {
		p.AddPeer(ctx, peer)
	}

	return p, nil
}

// AddPeer adds a peer to the Peers object.
func (p *Peers) AddPeer(ctx context.Context, peer string) error {
	p.peerMtx.Lock()
	defer p.peerMtx.Unlock()

	_, ok := p.whitelistPeers[peer]
	if ok {
		return nil
	}

	tx, err := p.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	p.whitelistPeers[peer] = true
	// Persist peers to disk
	_, err = tx.Execute(ctx, addPeer, peer)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// RemovePeer removes a peer from the Peers object.
func (p *Peers) RemovePeer(ctx context.Context, peer string) error {
	p.peerMtx.Lock()
	defer p.peerMtx.Unlock()

	_, ok := p.whitelistPeers[peer] // check if peer exists
	if !ok {
		return nil
	}

	tx, err := p.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	delete(p.whitelistPeers, peer)

	if p.removePeerFn != nil {
		if err := p.removePeerFn(peer); err != nil {
			return err
		}
	}

	// Persist peers to disk
	_, err = tx.Execute(ctx, removePeer, peer)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// HasPeer checks if a peer is in the Peers object.
func (p *Peers) HasPeer(ctx context.Context, peer string) bool {
	p.peerMtx.RLock()
	defer p.peerMtx.RUnlock()

	// Check if peer is in the whitelistPeers
	_, ok := p.whitelistPeers[peer]
	return ok
}

func (p *Peers) loadPeers(ctx context.Context) error {
	tx, err := p.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Execute(ctx, `SELECT peer_id FROM `+p2pSchemaName+`.peers;`)
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

func (p *Peers) SetRemovePeerFn(fn func(peerID string) error) {
	p.removePeerFn = fn
}

// NodeIDAddressString makes a full CometBFT node ID address string in the
// format <nodeID>@hostPort where nodeID is derived from the provided public
// key.
func NodeIDAddressString(pubkey ed25519.PubKey, hostPort string) string {
	nodeID := p2p.PubKeyToID(pubkey)
	return p2p.IDAddressString(nodeID, hostPort)
}
