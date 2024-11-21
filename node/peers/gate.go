package peers

import (
	"fmt"
	"slices"

	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

// ProtocolGater is a ConnectionGater for a libp2p Host. This was supposed to be
// used to ensure protocol support before allowing the connection to complete,
// but it seems that protocol negotiation occurs after this.  MAY REMOVE.
type ProtocolGater struct {
	ps                peerstore.Peerstore
	requiredProtocols []protocol.ID
}

func NewProtocolGater(requiredProtocols []protocol.ID) *ProtocolGater {
	return &ProtocolGater{requiredProtocols: requiredProtocols}
}

func (pg *ProtocolGater) SetPeerStore(ps peerstore.Peerstore) {
	pg.ps = ps
}

func (pg *ProtocolGater) InterceptPeerDial(p peer.ID) bool {
	return true
}

func (pg *ProtocolGater) InterceptAddrDial(id peer.ID, addr multiaddr.Multiaddr) bool {
	return true
}

func (pg *ProtocolGater) InterceptAccept(network.ConnMultiaddrs) bool {
	return true
}

func (pg *ProtocolGater) InterceptSecured(dir network.Direction, pid peer.ID, conn network.ConnMultiaddrs) bool {
	return true
}

func (pg *ProtocolGater) InterceptUpgraded(conn network.Conn) (bool, control.DisconnectReason) {
	if pg.ps == nil {
		return true, 0
	}
	peerID := conn.RemotePeer()
	// Check if the peer supports the required protocols
	supportedProtos, err := pg.ps.GetProtocols(peerID)
	if err != nil {
		return false, 0
	}
	fmt.Printf("protos supported by %v: %v\n", peerID, supportedProtos)

	for _, protoID := range pg.requiredProtocols {
		if !slices.Contains(supportedProtos, protoID) {
			return false, 0
		}
	}
	return true, 0
	// if err := RequirePeerProtos(context.Background(), pg.ps, peerID, pg.requiredProtocols...); err != nil {
	// 	fmt.Printf("Disconnecting peer %s due to missing protocols: %v\n", peerID, err)
	// 	return false, 0 // Reject the connection
	// }
	// return true, 0
}
