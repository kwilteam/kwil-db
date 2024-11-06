package peers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"kwil/log"
	"kwil/node/types"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/multiformats/go-multiaddr"
)

const (
	maxRetries      = 500
	baseDelay       = 2 * time.Second
	disconnectLimit = 7 * 24 * time.Hour // 1 week
)

type Connector interface {
	Connect(ctx context.Context, pi peer.AddrInfo) error
	// ClosePeer(peer.ID) error
	// Connectedness(peer.ID) network.Connectedness
	// LocalPeer() peer.ID
}

// peer/connection manager:
//	1. manage peerstore (and it's address book)
//  2. store and load from our address book file
//	3. provide the Notifee (connect/disconnect hooks)
//	4. maintain connections (min/max limits)

type RemotePeersFn func(ctx context.Context, peerID peer.ID) ([]peer.AddrInfo, error)

type PeerMan struct {
	log log.Logger
	h   host.Host // TODO: remove with just interfaces needed
	c   Connector
	ps  peerstore.Peerstore

	requestPeers RemotePeersFn

	pex      bool
	addrBook string

	done  chan struct{}
	close func()
	wg    sync.WaitGroup

	// TODO: revise address book file format as needed if these should persist
	mtx         sync.Mutex
	disconnects map[peer.ID]time.Time // Track disconnection timestamps
}

func NewPeerMan(pex bool, addrBook string, logger log.Logger, h host.Host,
	requestPeers RemotePeersFn) (*PeerMan, error) {
	if logger == nil {
		logger = log.DiscardLogger
	}
	done := make(chan struct{})
	pm := &PeerMan{
		h:    h, // tmp
		c:    h,
		ps:   h.Peerstore(),
		log:  logger,
		done: done,
		close: sync.OnceFunc(func() {
			close(done)
		}),
		pex:          pex,
		requestPeers: requestPeers,
		addrBook:     addrBook,
		disconnects:  make(map[peer.ID]time.Time),
	}

	peerInfo, err := loadPeers(pm.addrBook)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to load address book %s", pm.addrBook)
	}
	numPeers := pm.addPeers(peerInfo, peerstore.RecentlyConnectedAddrTTL)
	logger.Infof("Loaded address book with %d peers", numPeers)

	return pm, nil
}

var _ discovery.Discoverer = (*PeerMan)(nil) // FindPeers method

func (pm *PeerMan) Start(ctx context.Context) error {
	if pm.pex {
		pm.wg.Add(1)
		go func() {
			defer pm.wg.Done()
			pm.startPex(ctx)
		}()
	}

	pm.wg.Add(1)
	go func() {
		defer pm.wg.Done()
		pm.removeOldPeers()
	}()

	<-ctx.Done()

	pm.close()

	pm.wg.Wait()

	return nil
}

func (pm *PeerMan) startPex(ctx context.Context) {
	for {
		// discover for this node
		peerChan, err := pm.FindPeers(ctx, "kwil_namespace")
		if err != nil {
			pm.log.Errorf("FindPeers: %v", err)
		} else {
			go func() {
				var count int
				for peer := range peerChan {
					if pm.addPeerAddrs(peer) {
						// TODO: connection manager, with limits
						if err = pm.c.Connect(ctx, peer); err != nil {
							pm.log.Warnf("Failed to connect to %s: %v", peer.ID, err)
						}
					}
					count++
				}
				if count > 0 {
					if err := pm.savePeers(); err != nil {
						pm.log.Warnf("Failed to write address book: %v", err)
					}
				}
			}()
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(20 * time.Second):
		}

		if err := pm.savePeers(); err != nil {
			pm.log.Warnf("Failed to write address book: %v", err)
		}
	}
}

func (pm *PeerMan) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	peerChan := make(chan peer.AddrInfo)

	peers := pm.h.Network().Peers()
	if len(peers) == 0 {
		close(peerChan)
		pm.log.Warn("no existing peers for peer discovery")
		return peerChan, nil
	}

	var wg sync.WaitGroup
	wg.Add(len(peers))
	for _, peerID := range peers {
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			peers, err := pm.requestPeers(ctx, peerID)
			if err != nil {
				pm.log.Warnf("Failed to get peers from %v: %v", peerID, err)
				return
			}

			for _, p := range peers {
				peerChan <- p
			}
		}()
	}

	go func() {
		wg.Wait()
		close(peerChan)
	}()

	return peerChan, nil
}

func (pm *PeerMan) KnownPeers() []types.PeerInfo {
	var peers []types.PeerInfo
	for _, peerID := range pm.h.Network().Peers() { // connected peers only
		addrs := pm.ps.Addrs(peerID)

		supportedProtos, err := pm.ps.GetProtocols(peerID)
		if err != nil {
			pm.log.Errorf("GetProtocols for %v: %v", peerID, err)
			continue
		}

		peers = append(peers, types.PeerInfo{
			AddrInfo: types.AddrInfo{
				ID:    peerID,
				Addrs: addrs,
			},
			Protos: supportedProtos,
		})

	}
	return peers
}

func (pm *PeerMan) PrintKnownPeers() {
	peers := pm.KnownPeers()
	for _, p := range peers {
		pm.log.Info("Known peer", "id", p.ID.String())
	}
}

func (pm *PeerMan) savePeers() error {
	peerList := pm.KnownPeers()
	if err := persistPeers(peerList, pm.addrBook); err != nil {
		return err
	}
	return nil
}

// persistPeers saves known peers to a JSON file
func persistPeers(peers []types.PeerInfo, filePath string) error {
	// Marshal peerList to JSON
	data, err := json.MarshalIndent(peers, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling peers to JSON: %v", err)
	}

	// Write to file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("writing peers to file: %w", err)
	}
	return nil
}

func loadPeers(filePath string) ([]types.PeerInfo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read peerstore file: %w", err)
	}

	var peerList []types.PeerInfo
	if err := json.Unmarshal(data, &peerList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal peerstore data: %w", err)
	}
	return peerList, nil
}

func (pm *PeerMan) addPeers(peerList []types.PeerInfo, ttl time.Duration) int {
	var count int
	for _, pInfo := range peerList {
		// for _, addr := range pInfo.Addrs {
		// 	ps.AddAddr(pInfo.ID, addr, peerstore.PermanentAddrTTL)
		// }
		addrs := pm.ps.Addrs(pInfo.ID)
		for _, addr := range pInfo.Addrs {
			if !multiaddr.Contains(addrs, addr) {
				pm.ps.AddAddr(pInfo.ID, addr, ttl)
				pm.log.Infof("Added new peer address to store: %v @ %v", pInfo.ID, addr)
				count++
			}
			// TODO: we need a connect hook to change to forever on connect
		}
		//addPeerAddrs(ps, peer.AddrInfo(pInfo.AddrInfo))
		for _, proto := range pInfo.Protos {
			if err := pm.ps.AddProtocols(pInfo.ID, proto); err != nil {
				pm.log.Warnf("Error adding protocol %s for peer %s: %v", proto, pInfo.ID, err)
			}
		}
	}

	return count
}

// addPeerAddrs adds a discovered peer to the local peer store.
func (pm *PeerMan) addPeerAddrs(p peer.AddrInfo) (added bool) {
	numAdded := pm.addPeers([]types.PeerInfo{
		{
			AddrInfo: types.AddrInfo(p),
			// No known protocols, yet
		},
	}, peerstore.TempAddrTTL)
	return numAdded > 0
}

// Connected is triggered when a peer connects
func (pm *PeerMan) Connected(net network.Network, conn network.Conn) {
	peerID := conn.RemotePeer()
	pm.log.Infof("Connected to peer %s", peerID)

	// pm.ps.UpdateAddrs(peerID, ttlProvisional, ttlKnown)

	// Reset disconnect timestamp on successful connection
	pm.mtx.Lock()
	defer pm.mtx.Unlock()
	delete(pm.disconnects, peerID)
}

// Disconnected is triggered when a peer disconnects
func (pm *PeerMan) Disconnected(net network.Network, conn network.Conn) {
	peerID := conn.RemotePeer()
	pm.log.Infof("Disconnected from peer %v", peerID)
	// Store disconnection timestamp
	pm.mtx.Lock()
	defer pm.mtx.Unlock()
	pm.disconnects[peerID] = time.Now()

	select {
	case <-pm.done:
		return
	default:
	}

	// Create a context that is canceled if pm.done is closed (Notifiee
	// interface doesn't pass one).
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		select {
		case <-ctx.Done():
			return
		case <-pm.done:
			return
		}
	}()

	pm.wg.Add(1)
	go func() {
		defer pm.wg.Done()
		defer cancel()
		pm.reconnectWithRetry(ctx, peerID)
	}()
}

func (pm *PeerMan) Listen(network.Network, multiaddr.Multiaddr)      {}
func (pm *PeerMan) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Reconnect logic with retry

// Reconnect logic with exponential backoff and capped retries
func (pm *PeerMan) reconnectWithRetry(ctx context.Context, peerID peer.ID) {
	for attempt := range maxRetries {
		addrInfo := peer.AddrInfo{
			ID:    peerID,
			Addrs: pm.ps.Addrs(peerID),
		}

		delay := baseDelay * (1 << attempt)
		if delay > 1*time.Minute {
			delay = 1 * time.Minute // Cap delay at 1 minute
		}

		pm.log.Infof("Attempting reconnection to peer %s (attempt %d/%d)", peerID, attempt+1, maxRetries)
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		if err := pm.c.Connect(ctx, addrInfo); err != nil {
			cancel()
			var dErr *swarm.DialError
			if errors.Is(err, swarm.ErrAllDialsFailed) && errors.As(err, &dErr) {
				// the actual DialError string is multi-line
				addrs := make([]string, len(dErr.DialErrors))
				for i, te := range dErr.DialErrors {
					addrs[i] = te.Address.String()
				}
				err = fmt.Errorf("%w: [%s]", dErr.Cause, strings.Join(addrs, ", "))
			}
			pm.log.Infof("Failed to reconnect to peer %s (trying again in %v): %v", peerID, delay, err)

		} else {
			cancel()
			pm.log.Infof("Successfully reconnected to peer %s", peerID)
			return
		}

		select {
		case <-pm.done:
			return
		case <-time.After(delay):
		}
	}
	pm.log.Infof("Exceeded max retries for peer %s. Giving up.", peerID)
}

// Periodically remove peers disconnected for over a week
func (pm *PeerMan) removeOldPeers() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-pm.done:
			return
		case <-ticker.C:
		}

		now := time.Now()
		func() {
			pm.mtx.Lock()
			defer pm.mtx.Unlock()
			for peerID, disconnectTime := range pm.disconnects {
				if now.Sub(disconnectTime) > disconnectLimit {
					pm.ps.RemovePeer(peerID)
					delete(pm.disconnects, peerID) // Remove from tracking map
					pm.log.Infof("Removed peer %s last connected %v ago", peerID, time.Since(disconnectTime))
				}
			}
		}()
	}
}
