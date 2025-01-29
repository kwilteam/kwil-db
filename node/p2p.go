package node

import (
	"context"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strconv"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/node/peers"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
)

type P2PService struct {
	pm        peerManager // *peers.PeerMan
	host      host.Host
	dht       *dht.IpfsDHT
	discovery discovery.Discovery

	log log.Logger
}

type P2PServiceConfig struct {
	PrivKey crypto.PrivateKey
	RootDir string
	ChainID string
	KwilCfg *config.Config

	Logger log.Logger
}

func NewP2PService(cfg *P2PServiceConfig, host host.Host) (*P2PService, error) {
	// This connection gater is logically be part of PeerMan, but the libp2p
	// Host constructor needs it, and PeerMan needs Host for its peerstore
	// and connect method. For now we create it here and give it to both.
	logger := cfg.Logger

	var wcg *peers.WhitelistGater
	if cfg.KwilCfg.P2P.PrivateMode {
		logger.Infof("Private P2P mode enabled")
		var peerWhitelist []peer.ID
		for _, nodeID := range cfg.KwilCfg.P2P.Whitelist {
			peerID, err := nodeIDToPeerID(nodeID)
			if err != nil {
				return nil, fmt.Errorf("invalid whitelist node ID: %w", err)
			}
			peerWhitelist = append(peerWhitelist, peerID)
			logger.Infof("Adding peer to whitelist: %v", nodeID)
		}
		wcg = peers.NewWhitelistGater(peerWhitelist, peers.WithLogger(logger.New("PEERFILT")))
		// PeerMan adds more from address book.
	}
	cg := peers.ChainConnectionGaters(wcg) // pointless for one, but can be more

	if host == nil {
		ip, portStr, err := net.SplitHostPort(cfg.KwilCfg.P2P.ListenAddress)
		if err != nil {
			return nil, fmt.Errorf("invalid P2P listen address: %w", err)
		}
		if ip == "" { // handle ":6600" to mean listen on all interfaces
			ip = "0.0.0.0"
		}

		port, err := strconv.ParseUint(portStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid P2P listen port: %s, %w", portStr, err)
		}

		hostCfg := &hostConfig{
			ip:              ip,
			port:            port,
			privKey:         cfg.PrivKey,
			chainID:         cfg.ChainID,
			connGater:       cg,
			logger:          logger,
			externalAddress: cfg.KwilCfg.P2P.ExternalAddress,
		}

		host, err = newHost(hostCfg)
		if err != nil {
			return nil, fmt.Errorf("cannot create host: %w", err)
		}
	}

	addrBookPath := filepath.Join(cfg.RootDir, "addrbook.json")

	pmCfg := &peers.Config{
		PEX:               cfg.KwilCfg.P2P.Pex,
		AddrBook:          addrBookPath,
		Logger:            logger.New("PEERS"),
		Host:              host,
		ChainID:           cfg.ChainID,
		TargetConnections: 20,
		ConnGater:         wcg,
		RequiredProtocols: RequiredStreamProtocols,
	}
	pm, err := peers.NewPeerMan(pmCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer manager: %w", err)
	}

	mode := dht.ModeServer
	ctx := context.Background()
	dht, err := makeDHT(ctx, host, nil, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}
	discoverer := makeDiscovery(dht)

	return &P2PService{
		pm:        pm,
		host:      host,
		dht:       dht,
		discovery: discoverer,
		log:       logger,
	}, nil
}

func (p *P2PService) Start(ctx context.Context, bootpeers ...string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p.host.Network().Notify(p.pm)
	defer p.host.Network().StopNotify(p.pm)

	bootpeersMA, err := peers.ConvertPeersToMultiAddr(bootpeers)
	if err != nil {
		return err
	}

	// connect to bootstrap peers, if any.
	//
	// NOTE: it may be preferable to simply add to Host's peer store here and
	// let PeerMan manage connections.
	for i, peer := range bootpeersMA {
		peerInfo, err := makePeerAddrInfo(peer)
		if err != nil {
			p.log.Warnf("invalid bootnode address %v from setting %v", peer, bootpeers[i])
			continue
		}

		// Don't dial ourself.
		if peerInfo.ID == p.host.ID() {
			continue
		}

		p.pm.Allow(peerInfo.ID)

		err = p.pm.Connect(ctx, peers.AddrInfo(*peerInfo))
		if err != nil {
			p.log.Errorf("failed to connect to %v: %v", bootpeers[i], err)
			// Add it to the peer store anyway since this was specified as a
			// bootnode, which is supposed to be persistent, so we should try to
			// connect again later.
			p.host.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)
			continue
		}
		p.log.Infof("Connected to bootstrap peer %v", bootpeers[i])
	} // else would use persistent peer store (address book)

	return nil
}

func (p *P2PService) Close() error {
	p.log.Info("Stopping P2P services...")
	var err error

	if err1 := p.dht.Close(); err1 != nil {
		p.log.Warn("Failed to cleanly stop the DHT service: %v", err1)
		err = errors.Join(err, fmt.Errorf("failed to stop DHT: %w", err1))
	}

	if err1 := p.host.Close(); err1 != nil {
		p.log.Warn("Failed to cleanly stop P2P host: %v", err1)
		err = errors.Join(err, fmt.Errorf("failed to stop host: %w", err1))
	}

	return err
}

func (p *P2PService) Host() host.Host {
	return p.host
}

func (p *P2PService) Discovery() discovery.Discovery {
	return p.discovery
}
