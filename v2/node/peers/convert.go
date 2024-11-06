package peers

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"kwil/crypto"

	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/multiformats/go-multiaddr"
)

// PeerIDFromPubKey converts a pubkey to a peer ID string.
func PeerIDFromPubKey(pubkey crypto.PublicKey) (string, error) {
	_, peerID, err := convertPubKey(pubkey)
	if err != nil {
		return "", err
	}
	return peerID.String(), nil // base58 encoding of identity multihash
}

// PubKeyFromPeerID tries to decode the pubkey from a peer ID string.
// This will only work if the peer ID is an "identity" multihash.
func PubKeyFromPeerID(peerID string) (crypto.PublicKey, error) {
	pid, err := peer.Decode(peerID)
	if err != nil {
		return nil, err
	}
	return pubKeyFromPeerID(pid)
}

func pubKeyFromPeerID(peerID peer.ID) (crypto.PublicKey, error) {
	p2pPubKey, err := peerID.ExtractPublicKey()
	if err != nil {
		return nil, err
	}
	rawPub, err := p2pPubKey.Raw()
	if err != nil {
		return nil, err
	}
	switch p2pPubKey.(type) {
	case *p2pcrypto.Secp256k1PublicKey:
		return crypto.UnmarshalSecp256k1PublicKey(rawPub)
	case *p2pcrypto.Ed25519PublicKey:
		return crypto.UnmarshalEd25519PublicKey(rawPub)
	default:
		return nil, fmt.Errorf("unsupported pubkey type: %T", p2pPubKey)
	}
}

func convertPubKey(pubkey crypto.PublicKey) (p2pcrypto.PubKey, peer.ID, error) {
	rawPub := pubkey.Bytes()
	var p2pPub p2pcrypto.PubKey
	var err error
	switch pubkey.(type) {
	case *crypto.Secp256k1PublicKey:
		p2pPub, err = p2pcrypto.UnmarshalSecp256k1PublicKey(rawPub)
	case *crypto.Ed25519PublicKey:
		p2pPub, err = p2pcrypto.UnmarshalEd25519PublicKey(rawPub)
	default:
		return nil, "", fmt.Errorf("unsupported pubkey type: %T", pubkey)
	}
	if err != nil {
		return nil, "", err
	}
	p2pAddr, err := peer.IDFromPublicKey(p2pPub)
	if err != nil {
		return nil, "", err
	}
	return p2pPub, p2pAddr, nil
}

// ConvertPeersToMultiAddr convert a peer from pubkeyHex#keyTypeInt@ip:port to /ip4/ip/tcp/port/p2p/peerID
func ConvertPeersToMultiAddr(peers []string) ([]string, error) {
	addrs := make([]string, len(peers))
	for i, peerAddr := range peers {
		// split the pieces of pubkey#type@ip:port
		parts := strings.Split(peerAddr, "@")
		if len(parts) != 2 {
			return nil, errors.New("invalid peer notation")
		}
		addr := parts[1]

		parts = strings.Split(parts[0], "#")
		if len(parts) != 2 {
			return nil, errors.New("invalid peer notation")
		}
		pubkeyStr, keyTypeStr := parts[0], parts[1]
		keyType, err := strconv.ParseUint(keyTypeStr, 10, 16)
		if err != nil {
			return nil, errors.New("invalid key type in peer notation")
		}
		pubkeyBts, err := hex.DecodeString(pubkeyStr)
		if err != nil {
			return nil, err
		}
		var pubkey crypto.PublicKey
		switch crypto.KeyType(keyType) {
		case crypto.KeyTypeSecp256k1:
			pubkey, err = crypto.UnmarshalSecp256k1PublicKey(pubkeyBts)
		case crypto.KeyTypeEd25519:
			pubkey, err = crypto.UnmarshalEd25519PublicKey(pubkeyBts)
		default:
			return nil, errors.New("unsupported key type")
		}
		if err != nil {
			return nil, err
		}
		peerID, err := PeerIDFromPubKey(pubkey)
		if err != nil {
			return nil, err
		}

		host, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return nil, err
		}

		maStr := fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s", host, port, peerID)
		// ensure the multiaddress string is parsable
		_, err = multiaddr.NewMultiaddr(maStr)
		if err != nil {
			return nil, err
		}
		addrs[i] = maStr
	}
	return addrs, nil
}

func CompressDialError(err error) error {
	var dErr *swarm.DialError
	if errors.Is(err, swarm.ErrAllDialsFailed) && errors.As(err, &dErr) {
		// the actual DialError string is multi-line
		addrs := make([]string, len(dErr.DialErrors))
		for i, te := range dErr.DialErrors {
			addrs[i] = te.Address.String()
		}
		err = fmt.Errorf("%w: [%s]", dErr.Cause, strings.Join(addrs, ", "))
	}
	return err
}
