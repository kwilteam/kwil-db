package peers

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

type AddrInfo struct {
	ID    peer.ID               `json:"id"`
	Addrs []multiaddr.Multiaddr `json:"addrs"`
}

type PeerInfo struct {
	AddrInfo
	Protos []protocol.ID `json:"protos"`
}

func (p PeerInfo) MarshalJSON() ([]byte, error) {
	var addrStrs []string
	for _, addr := range p.Addrs {
		addrStrs = append(addrStrs, addr.String())
	}
	var protoStrs []string
	for _, proto := range p.Protos {
		protoStrs = append(protoStrs, string(proto))
	}
	return json.Marshal(struct {
		ID     string   `json:"id"`
		Addrs  []string `json:"addrs"`
		Protos []string `json:"protos"`
	}{
		ID:     p.ID.String(),
		Addrs:  addrStrs,
		Protos: protoStrs,
	})
}

func (p *PeerInfo) UnmarshalJSON(data []byte) error {
	aux := struct {
		ID     string   `json:"id"`
		Addrs  []string `json:"addrs"`
		Protos []string `json:"protos"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	peerID, err := peer.Decode(aux.ID)
	if err != nil {
		return err
	}
	p.ID = peerID

	for _, addrStr := range aux.Addrs {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		p.Addrs = append(p.Addrs, addr)
	}
	for _, protoStr := range aux.Protos {
		p.Protos = append(p.Protos, protocol.ID(protoStr))
	}
	return nil
}

type PersistentPeerInfo struct {
	NodeID      string                `json:"id"` // "node ID" (pubkeybytes#keytype)
	Addrs       []multiaddr.Multiaddr `json:"addrs"`
	Protos      []protocol.ID         `json:"protos"`
	Whitelisted bool                  `json:"whitelisted"`
	// We probably need a last connected time and/or ttl
}

func (p PersistentPeerInfo) MarshalJSON() ([]byte, error) {
	var addrStrs []string
	for _, addr := range p.Addrs {
		addrStrs = append(addrStrs, addr.String())
	}
	var protoStrs []string
	for _, proto := range p.Protos {
		protoStrs = append(protoStrs, string(proto))
	}
	return json.Marshal(struct {
		ID          string   `json:"id"`
		Addrs       []string `json:"addrs"`
		Protos      []string `json:"protos"`
		Whitelisted bool     `json:"whitelisted"`
	}{
		ID:          p.NodeID,
		Addrs:       addrStrs,
		Protos:      protoStrs,
		Whitelisted: p.Whitelisted,
	})
}

func (p *PersistentPeerInfo) UnmarshalJSON(data []byte) error {
	aux := struct {
		ID          string   `json:"id"`
		Addrs       []string `json:"addrs"`
		Protos      []string `json:"protos"`
		Whitelisted bool     `json:"whitelisted"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.NodeID = aux.ID
	p.Whitelisted = aux.Whitelisted

	for _, addrStr := range aux.Addrs {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		p.Addrs = append(p.Addrs, addr)
	}
	for _, protoStr := range aux.Protos {
		p.Protos = append(p.Protos, protocol.ID(protoStr))
	}
	return nil
}
