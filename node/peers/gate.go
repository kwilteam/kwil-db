package peers

import (
	"sync"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// WhitelistGater is a libp2p connmgr.ConnectionGater implementation to enforce
// a peer whitelist.
type WhitelistGater struct {
	logger log.Logger

	mtx       sync.RWMutex // very infrequent whitelist updates
	permitted map[peer.ID]bool
}

type gateOpts struct {
	logger log.Logger
}

type GateOpt func(*gateOpts)

func WithLogger(logger log.Logger) GateOpt {
	return func(opts *gateOpts) {
		opts.logger = logger
	}
}

func NewWhitelistGater(allowed []peer.ID, opts ...GateOpt) *WhitelistGater {
	options := &gateOpts{
		logger: log.DiscardLogger,
	}
	for _, opt := range opts {
		opt(options)
	}

	permitted := make(map[peer.ID]bool)
	for _, pid := range allowed {
		permitted[pid] = true
	}

	return &WhitelistGater{
		logger:    options.logger,
		permitted: permitted,
	}
}

// Allow and Disallow work with a nil *WhitelistGater, but not the
// connmgr.ConnectionGater methods. So, do not give a nil *WhitelistGater to
// libp2p.New via libp2p.ConnectionGater.

// Allow adds a peer to the whitelist.
func (g *WhitelistGater) Allow(p peer.ID) {
	if g == nil {
		return
	}
	g.mtx.Lock()
	defer g.mtx.Unlock()
	g.permitted[p] = true
}

// Disallow removes a peer from the whitelist.
func (g *WhitelistGater) Disallow(p peer.ID) {
	if g == nil {
		return
	}
	g.mtx.Lock()
	defer g.mtx.Unlock()
	delete(g.permitted, p)
}

// Allowed returns the list of peers in the whitelist.
func (g *WhitelistGater) Allowed() []peer.ID {
	if g == nil {
		return nil
	}
	g.mtx.RLock()
	defer g.mtx.RUnlock()
	allowed := make([]peer.ID, 0, len(g.permitted))
	for pid := range g.permitted {
		allowed = append(allowed, pid)
	}
	return allowed
}

// IsAllowed indicates if a peer is in the whitelist. This is mainly for the
// connmgr.ConnectionGater methods.
func (g *WhitelistGater) IsAllowed(p peer.ID) bool {
	if g == nil {
		return true
	}
	g.mtx.RLock()
	defer g.mtx.RUnlock()
	return g.permitted[p]
}

var _ connmgr.ConnectionGater = (*WhitelistGater)(nil)

// OUTBOUND

func (g *WhitelistGater) InterceptPeerDial(p peer.ID) bool {
	ok := g.IsAllowed(p)
	if !ok {
		g.logger.Infof("Blocking OUTBOUND dial to peer not on whitelist: %v", p)
	}
	return ok
}

func (g *WhitelistGater) InterceptAddrDial(p peer.ID, addr multiaddr.Multiaddr) bool {
	// InterceptPeerDial came first, don't bother doing it again here. Only
	// filter here if we want to filter by network address.
	return true
}

// INBOUND

func (g *WhitelistGater) InterceptAccept(connAddrs network.ConnMultiaddrs) bool {
	// Filter here if we want to filter by network address; we get the peer ID
	// after a secure connection is established (InterceptSecured).
	return true
}

func (g *WhitelistGater) InterceptSecured(dir network.Direction, p peer.ID, conn network.ConnMultiaddrs) bool {
	ok := g.IsAllowed(p)
	if !ok {
		g.logger.Infof("Blocking INBOUND connection from peer not on whitelist: %v", p)
	}
	return ok
}

func (g *WhitelistGater) InterceptUpgraded(conn network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}
