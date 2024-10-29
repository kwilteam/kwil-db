package node

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"time"

	"p2p/node/types"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	blkReadLimit  = 300_000_000
	blkGetTimeout = 90 * time.Second
)

func (n *Node) blkGetStreamHandler(s network.Stream) {
	defer s.Close()

	var req blockHashReq
	if _, err := req.ReadFrom(s); err != nil {
		fmt.Println("bad get blk request:", err)
		return
	}

	blk, appHash, err := n.bki.Get(req.Hash)
	if err != nil {
		s.Write(noData) // don't have it
	} else {
		rawBlk := types.EncodeBlock(blk)
		binary.Write(s, binary.LittleEndian, blk.Header.Height)
		s.Write(appHash[:])
		s.Write(rawBlk)
	}
}

func (n *Node) blkGetHeightStreamHandler(s network.Stream) {
	defer s.Close()

	var req blockHeightReq
	if _, err := req.ReadFrom(s); err != nil {
		fmt.Println("bad get blk request:", err)
		return
	}
	log.Printf("requested block height: %d", req.Height)

	hash, blk, appHash, err := n.bki.GetByHeight(req.Height)
	if err != nil {
		s.Write(noData) // don't have it
	} else {
		rawBlk := types.EncodeBlock(blk) // blkHash := blk.Hash()
		// maybe we remove hash from the protocol, was thinking receiver could
		// hang up earlier depending...
		s.Write(hash[:])
		s.Write(appHash[:])
		s.Write(rawBlk)
	}
}

func (n *Node) blkAnnStreamHandler(s network.Stream) {
	defer s.Close()

	s.SetDeadline(time.Now().Add(blkGetTimeout))
	ctx, cancel := context.WithTimeout(context.Background(), blkGetTimeout)
	defer cancel()

	var reqMsg blockInitMsg
	if _, err := reqMsg.ReadFrom(s); err != nil {
		log.Println("bad blk ann request:", err)
		return
	}

	height, blkHash, appHash := reqMsg.Height, reqMsg.Hash, reqMsg.AppHash
	blkid := blkHash.String()

	if height < 0 {
		log.Printf("invalid height in blk ann request: %d", height)
		return
	}
	log.Printf("blk announcement received: %q / %d", blkid, height)

	// If we are a validator and this is the commit ann for a proposed block
	// that we already started executing, consensus engine will handle it.
	if !n.ce.AcceptCommit(height, blkHash, appHash) {
		return
	}

	// Possibly ce will handle it regardless.  For now, below is block store
	// code like a sentry node might do.

	need, done := n.bki.PreFetch(blkHash)
	defer done()
	if !need {
		return // we have or are currently fetching it, do nothing, assuming we have already re-announced
	}

	log.Printf("retrieving new block: %q", blkid)
	t0 := time.Now()

	// First try to get from this stream.
	rawBlk, err := request(s, []byte(getMsg), blkReadLimit)
	if err != nil {
		log.Printf("announcer failed to provide %v, trying other peers", blkid)
		// Since we are aware, ask other peers. we could also put this in a goroutine
		s.Close() // close the announcers stream first
		var gotHeight int64
		var gotAppHash types.Hash
		gotHeight, rawBlk, gotAppHash, err = n.getBlkWithRetry(ctx, blkHash, 500*time.Millisecond, 10)
		if err != nil {
			log.Printf("unable to retrieve tx %v: %v", blkid, err)
			return
		}
		if gotHeight != height {
			log.Printf("getblk response had unexpected height: wanted %d, got %d", height, gotHeight)
			return
		}
		if gotAppHash != appHash {
			log.Printf("getblk response had unexpected appHash: wanted %v, got %v", appHash, gotAppHash)
			return
		}
	}

	log.Printf("obtained content for block %q in %v", blkid, time.Since(t0))

	blk, err := types.DecodeBlock(rawBlk)
	if err != nil {
		log.Printf("decodeBlock failed for %v: %v", blkid, err)
		return
	}
	if blk.Header.Height != height {
		log.Printf("getblk response had unexpected height: wanted %d, got %d", height, blk.Header.Height)
		return
	}
	gotBlkHash := blk.Header.Hash()
	if gotBlkHash != blkHash {
		log.Printf("invalid block hash: wanted %v, got %x", blkHash, gotBlkHash)
		return
	}

	// re-announce

	go func() {
		n.ce.NotifyBlockCommit(blk, appHash)
		n.announceRawBlk(context.Background(), blkHash, height, rawBlk, appHash, s.Conn().RemotePeer())
	}()
}

func (n *Node) announceBlk(ctx context.Context, blk *types.Block, appHash types.Hash, from peer.ID) {
	blkHash := blk.Hash()
	fmt.Println("announceBlk", blk.Header.Height, blkHash, appHash, from)
	rawBlk := types.EncodeBlock(blk)
	n.announceRawBlk(ctx, blkHash, blk.Header.Height, rawBlk, appHash, from)
	return
}

func (n *Node) announceRawBlk(ctx context.Context, blkHash types.Hash, height int64,
	rawBlk []byte, appHash types.Hash, from peer.ID) {
	peers := n.peers()
	if len(peers) == 0 {
		log.Println("no peers to advertise block to")
		return
	}

	for _, peerID := range peers {
		if peerID == from {
			continue
		}
		log.Printf("advertising block %s (height %d / txs %d) to peer %v",
			blkHash, height, len(rawBlk), peerID)
		resID, err := blockInitMsg{
			Hash:    blkHash,
			Height:  height,
			AppHash: appHash,
		}.MarshalBinary()
		if err != nil {
			log.Println(err)
			continue
		}
		err = advertiseToPeer(ctx, n.host, peerID, ProtocolIDBlkAnn,
			contentAnn{cType: "block announce", ann: resID, content: rawBlk})
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func (n *Node) getBlkWithRetry(ctx context.Context, blkHash types.Hash, baseDelay time.Duration,
	maxAttempts int) (int64, []byte, types.Hash, error) {
	var attempts int
	for {
		height, raw, appHash, err := n.getBlk(ctx, blkHash)
		if err == nil {
			return height, raw, appHash, nil
		}

		log.Printf("unable to retrieve block %v (%v), waiting to retry", blkHash, err)

		select {
		case <-ctx.Done():
		case <-time.After(baseDelay):
		}
		baseDelay *= 2
		attempts++
		if attempts >= maxAttempts {
			return 0, nil, types.Hash{}, ErrBlkNotFound
		}
	}
}

func (n *Node) getBlk(ctx context.Context, blkHash types.Hash) (int64, []byte, types.Hash, error) {
	for _, peer := range n.peers() {
		log.Printf("requesting block %v from %v", blkHash, peer)
		t0 := time.Now()
		resID, _ := blockHashReq{Hash: blkHash}.MarshalBinary()
		resp, err := requestFrom(ctx, n.host, peer, resID, ProtocolIDBlock, blkReadLimit)
		if errors.Is(err, ErrNotFound) {
			log.Printf("block not available on %v", peer)
			continue
		}
		if errors.Is(err, ErrNoResponse) {
			log.Printf("no response to block request to %v", peer)
			continue
		}
		if err != nil {
			log.Printf("unexpected error from %v: %v", peer, err)
			continue
		}

		if len(resp) < 8 {
			log.Printf("block response too short")
			continue
		}

		log.Printf("obtained content for block %q in %v", blkHash, time.Since(t0))

		height := binary.LittleEndian.Uint64(resp[:8])
		var appHash types.Hash
		copy(appHash[:], resp[8:8+types.HashLen])
		rawBlk := resp[8+types.HashLen:]

		return int64(height), rawBlk, appHash, nil
	}
	return 0, nil, types.Hash{}, ErrBlkNotFound
}

func (n *Node) getBlkHeight(ctx context.Context, height int64) (types.Hash, types.Hash, []byte, error) {
	for _, peer := range n.peers() {
		log.Printf("requesting block number %d from %v", height, peer)
		t0 := time.Now()
		resID, _ := blockHeightReq{Height: height}.MarshalBinary()
		resp, err := requestFrom(ctx, n.host, peer, resID, ProtocolIDBlockHeight, blkReadLimit)
		if errors.Is(err, ErrNotFound) {
			log.Printf("block not available on %v", peer)
			continue
		}
		if errors.Is(err, ErrNoResponse) {
			log.Printf("no response to block request to %v", peer)
			continue
		}
		if err != nil {
			log.Printf("unexpected error from %v: %v", peer, err)
			continue
		}

		if len(resp) < types.HashLen+1 {
			log.Printf("block response too short")
			continue
		}

		log.Printf("obtained content for block number %d in %v", height, time.Since(t0))

		var hash, appHash types.Hash
		copy(hash[:], resp[:types.HashLen])
		copy(appHash[:], resp[types.HashLen:types.HashLen*2])
		rawBlk := resp[types.HashLen*2:]

		return hash, appHash, rawBlk, nil
	}
	return types.Hash{}, types.ZeroHash, nil, ErrBlkNotFound
}
