package node

import (
	"context"
	"encoding/binary"
	"errors"
	"time"

	"kwil/node/types"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	blkReadLimit   = 300_000_000
	blkGetTimeout  = 90 * time.Second
	blkSendTimeout = 45 * time.Second
)

func (n *Node) blkGetStreamHandler(s network.Stream) {
	defer s.Close()

	var req blockHashReq
	if _, err := req.ReadFrom(s); err != nil {
		n.log.Warn("Bad get block (hash) request", "error", err) // Debug when we ship
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
		n.log.Warn("Bad get block (height) request", "error", err) // Debug when we ship
		return
	}
	n.log.Debug("Peer requested block", "height", req.Height)

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

	s.SetDeadline(time.Now().Add(blkGetTimeout + annRespTimeout + annWriteTimeout)) // combined
	ctx, cancel := context.WithTimeout(context.Background(), blkGetTimeout)
	defer cancel()

	var reqMsg blockAnnMsg
	if _, err := reqMsg.ReadFrom(s); err != nil {
		n.log.Warn("bad blk ann request", "error", err)
		return
	}

	height, blkHash, appHash, sig := reqMsg.Height, reqMsg.Hash, reqMsg.AppHash, reqMsg.LeaderSig
	blkid := blkHash.String()

	// TODO: also get and pass the signature to AcceptCommit to ensure it's
	// legit before we waste bandwidth on spam. We could also make the protocol
	// request the block header, and then CE checks block header.

	if height < 0 {
		n.log.Warn("invalid height in blk ann request", "height", height)
		return
	}
	n.log.Info("blk announcement received", "hash", blkid, "height", height)

	// If we are a validator and this is the commit ann for a proposed block
	// that we already started executing, consensus engine will handle it.
	if !n.ce.AcceptCommit(height, blkHash, appHash, sig) {
		return
	}

	// Possibly ce will handle it regardless.  For now, below is block store
	// code like a sentry node might do.

	need, done := n.bki.PreFetch(blkHash)
	defer done()
	if !need {
		return // we have or are currently fetching it, do nothing, assuming we have already re-announced
	}

	n.log.Info("retrieving new block", "hash", blkid)
	t0 := time.Now()

	// First try to get from this stream.
	rawBlk, err := request(s, []byte(getMsg), blkReadLimit)
	if err != nil {
		n.log.Warnf("announcer failed to provide %v, trying other peers", blkid)
		// Since we are aware, ask other peers. we could also put this in a goroutine
		s.Close() // close the announcers stream first
		var gotHeight int64
		var gotAppHash types.Hash
		gotHeight, rawBlk, gotAppHash, err = n.getBlkWithRetry(ctx, blkHash, 500*time.Millisecond, 10)
		if err != nil {
			n.log.Errorf("unable to retrieve tx %v: %v", blkid, err)
			return
		}
		if gotHeight != height {
			n.log.Errorf("getblk response had unexpected height: wanted %d, got %d", height, gotHeight)
			return
		}
		if gotAppHash != appHash {
			n.log.Errorf("getblk response had unexpected appHash: wanted %v, got %v", appHash, gotAppHash)
			return
		}
	}

	n.log.Infof("obtained content for block %q in %v", blkid, time.Since(t0))

	blk, err := types.DecodeBlock(rawBlk)
	if err != nil {
		n.log.Infof("decodeBlock failed for %v: %v", blkid, err)
		return
	}
	if blk.Header.Height != height {
		n.log.Infof("getblk response had unexpected height: wanted %d, got %d", height, blk.Header.Height)
		return
	}
	gotBlkHash := blk.Header.Hash()
	if gotBlkHash != blkHash {
		n.log.Infof("invalid block hash: wanted %v, got %x", blkHash, gotBlkHash)
		return
	}

	// re-announce

	go func() {
		n.ce.NotifyBlockCommit(blk, appHash)
		n.announceRawBlk(context.Background(), blkHash, height, rawBlk, appHash, s.Conn().RemotePeer(), reqMsg.LeaderSig) // re-announce with the leader's signature
	}()
}

func (n *Node) announceBlk(ctx context.Context, blk *types.Block, appHash types.Hash) {
	blkHash := blk.Hash()
	n.log.Debugln("announceBlk", blk.Header.Height, blkHash, appHash)
	rawBlk := types.EncodeBlock(blk)
	from := n.host.ID() // this announcement originates from us (not a reannouncement)
	n.announceRawBlk(ctx, blkHash, blk.Header.Height, rawBlk, appHash, from, blk.Signature)
}

func (n *Node) announceRawBlk(ctx context.Context, blkHash types.Hash, height int64,
	rawBlk []byte, appHash types.Hash, from peer.ID, blkSig []byte) {
	peers := n.peers()
	if len(peers) == 0 {
		n.log.Warn("No peers to advertise block to")
		return
	}

	for _, peerID := range peers {
		if peerID == from {
			continue
		}
		n.log.Infof("advertising block %s (height %d / sz %d) to peer %v",
			blkHash, height, len(rawBlk), peerID)
		resID, err := blockAnnMsg{
			Hash:      blkHash,
			Height:    height,
			AppHash:   appHash,
			LeaderSig: blkSig,
		}.MarshalBinary()
		if err != nil {
			n.log.Error("Unable to marshal block announcement", "error", err)
			continue
		}
		ann := contentAnn{cType: "block announce", ann: resID, content: rawBlk}
		err = n.advertiseToPeer(ctx, peerID, ProtocolIDBlkAnn, ann, blkSendTimeout)
		if err != nil {
			n.log.Warn("Failed to advertise block", "peer", peerID, "error", err)
			continue
		}
		n.log.Debugf("Advertised content %s to peer %s", ann, peerID)
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

		n.log.Warnf("unable to retrieve block %v (%v), waiting to retry", blkHash, err)

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
		n.log.Infof("requesting block %v from %v", blkHash, peer)
		t0 := time.Now()
		resID, _ := blockHashReq{Hash: blkHash}.MarshalBinary()
		resp, err := requestFrom(ctx, n.host, peer, resID, ProtocolIDBlock, blkReadLimit)
		if errors.Is(err, ErrNotFound) {
			n.log.Info("block not available", "peer", peer, "hash", blkHash)
			continue
		}
		if errors.Is(err, ErrNoResponse) {
			n.log.Info("no response to block request", "peer", peer, "hash", blkHash)
			continue
		}
		if err != nil {
			n.log.Info("block request failed unexpectedly", "peer", peer, "hash", blkHash)
			continue
		}

		if len(resp) < 8 {
			n.log.Info("block response too short", "peer", peer, "hash", blkHash)
			continue
		}

		n.log.Info("Obtained content for block", "block", blkHash, "elapsed", time.Since(t0))

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
		n.log.Infof("requesting block number %d from %v", height, peer)
		t0 := time.Now()
		resID, _ := blockHeightReq{Height: height}.MarshalBinary()
		resp, err := requestFrom(ctx, n.host, peer, resID, ProtocolIDBlockHeight, blkReadLimit)
		if errors.Is(err, ErrNotFound) {
			n.log.Warnf("block not available on %v", peer)
			continue
		}
		if errors.Is(err, ErrNoResponse) {
			n.log.Warnf("no response to block request to %v", peer)
			continue
		}
		if err != nil {
			n.log.Warnf("unexpected error from %v: %v", peer, err)
			continue
		}

		if len(resp) < types.HashLen+1 {
			n.log.Warnf("block response too short")
			continue
		}

		n.log.Info("obtained block contents", "height", height, "elapsed", time.Since(t0))

		var hash, appHash types.Hash
		copy(hash[:], resp[:types.HashLen])
		copy(appHash[:], resp[types.HashLen:types.HashLen*2])
		rawBlk := resp[types.HashLen*2:]

		return hash, appHash, rawBlk, nil
	}
	return types.Hash{}, types.ZeroHash, nil, ErrBlkNotFound
}
