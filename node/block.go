package node

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"time"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
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

	s.SetReadDeadline(time.Now().Add(reqRWTimeout))

	var req blockHashReq
	if _, err := req.ReadFrom(s); err != nil {
		n.log.Debug("Bad get block (hash) request", "error", err)
		return
	}
	n.log.Debug("Peer requested block", "hash", req.Hash)

	blk, ci, err := n.bki.Get(req.Hash)
	if err != nil || ci == nil {
		s.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		s.Write(noData) // don't have it
	} else {
		rawBlk := ktypes.EncodeBlock(blk)
		ciBytes, _ := ci.MarshalBinary()
		s.SetWriteDeadline(time.Now().Add(blkSendTimeout))
		binary.Write(s, binary.LittleEndian, blk.Header.Height)
		ktypes.WriteBytes(s, ciBytes)
		ktypes.WriteBytes(s, rawBlk)
	}
}

func (n *Node) blkGetHeightStreamHandler(s network.Stream) {
	defer s.Close()

	s.SetReadDeadline(time.Now().Add(reqRWTimeout))

	var req blockHeightReq
	if _, err := req.ReadFrom(s); err != nil {
		n.log.Warn("Bad get block (height) request", "error", err) // Debug when we ship
		return
	}
	n.log.Debug("Peer requested block", "height", req.Height)

	hash, blk, ci, err := n.bki.GetByHeight(req.Height)
	if err != nil || ci == nil {
		s.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		s.Write(noData) // don't have it
	} else {
		rawBlk := ktypes.EncodeBlock(blk) // blkHash := blk.Hash()
		ciBytes, _ := ci.MarshalBinary()
		// maybe we remove hash from the protocol, was thinking receiver could
		// hang up earlier depending...
		s.SetWriteDeadline(time.Now().Add(blkSendTimeout))
		s.Write(hash[:])
		ktypes.WriteBytes(s, ciBytes)
		ktypes.WriteBytes(s, rawBlk)
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

	height, blkHash, ci, sig := reqMsg.Height, reqMsg.Hash, reqMsg.CommitInfo, reqMsg.LeaderSig
	blkid := blkHash.String()

	// TODO: also get and pass the signature to AcceptCommit to ensure it's
	// legit before we waste bandwidth on spam. We could also make the protocol
	// request the block header, and then CE checks block header.

	if height < 0 {
		n.log.Warn("invalid height in blk ann request", "height", height)
		return
	}
	n.log.Debug("blk announcement received", "hash", blkid, "height", height)

	// If we are a validator and this is the commit ann for a proposed block
	// that we already started executing, consensus engine will handle it.
	if !n.ce.AcceptCommit(height, blkHash, ci, sig) {
		return
	}

	// Possibly ce will handle it regardless.  For now, below is block store
	// code like a sentry node might do.

	need, done := n.bki.PreFetch(blkHash)
	defer done()
	if !need {
		n.log.Debug("ALREADY HAVE OR FETCHING BLOCK")
		return // we have or are currently fetching it, do nothing, assuming we have already re-announced
	}

	n.log.Debug("retrieving new block", "hash", blkid)
	t0 := time.Now()

	// First try to get from this stream.
	rawBlk, err := request(s, []byte(getMsg), blkReadLimit)
	if err != nil {
		n.log.Warnf("announcer failed to provide %v, trying other peers", blkid)
		// Since we are aware, ask other peers. we could also put this in a goroutine
		s.Close() // close the announcers stream first
		var gotHeight int64
		var gotCI *ktypes.CommitInfo

		gotHeight, rawBlk, gotCI, err = n.getBlkWithRetry(ctx, blkHash, 500*time.Millisecond, 10)
		if err != nil {
			n.log.Errorf("unable to retrieve tx %v: %v", blkid, err)
			return
		}
		if gotHeight != height {
			n.log.Errorf("getblk response had unexpected height: wanted %d, got %d", height, gotHeight)
			return
		}
		if gotCI != nil && gotCI.AppHash != ci.AppHash {
			n.log.Errorf("getblk response had unexpected appHash: wanted %v, got %v", ci.AppHash, gotCI.AppHash)
			return
		}
	}

	n.log.Debugf("obtained content for block %q in %v", blkid, time.Since(t0))

	blk, err := ktypes.DecodeBlock(rawBlk)
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
	n.ce.NotifyBlockCommit(blk, ci)
	go func() {
		n.announceRawBlk(context.Background(), blkHash, height, rawBlk, ci, s.Conn().RemotePeer(), reqMsg.LeaderSig) // re-announce with the leader's signature
	}()
}

func (n *Node) announceBlk(ctx context.Context, blk *ktypes.Block, ci *ktypes.CommitInfo) {
	blkHash := blk.Hash()
	n.log.Debugln("announceBlk", blk.Header.Height, blkHash, ci.AppHash)
	rawBlk := ktypes.EncodeBlock(blk)
	from := n.host.ID() // this announcement originates from us (not a reannouncement)
	n.announceRawBlk(ctx, blkHash, blk.Header.Height, rawBlk, ci, from, blk.Signature)
}

func (n *Node) announceRawBlk(ctx context.Context, blkHash types.Hash, height int64,
	rawBlk []byte, ci *ktypes.CommitInfo, from peer.ID, blkSig []byte) {
	peers := n.peers()
	if len(peers) == 0 {
		n.log.Warn("No peers to advertise block to")
		return
	}

	for _, peerID := range peers {
		if peerID == from {
			continue
		}
		n.log.Infof("advertising block %s (height %d / sz %d / updates %v) to peer %v",
			blkHash, height, len(rawBlk), ci.ParamUpdates, peerID)
		resID, err := blockAnnMsg{
			Hash:       blkHash,
			Height:     height,
			CommitInfo: ci,
			LeaderSig:  blkSig,
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
	maxAttempts int) (int64, []byte, *ktypes.CommitInfo, error) {
	var attempts int
	for {
		height, raw, ci, err := n.getBlk(ctx, blkHash)
		if err == nil {
			return height, raw, ci, nil
		}

		n.log.Warnf("unable to retrieve block %v (%v), waiting to retry", blkHash, err)

		select {
		case <-ctx.Done():
		case <-time.After(baseDelay):
		}
		baseDelay *= 2
		attempts++
		if attempts >= maxAttempts {
			return 0, nil, nil, ErrBlkNotFound
		}
	}
}

func (n *Node) getBlk(ctx context.Context, blkHash types.Hash) (int64, []byte, *ktypes.CommitInfo, error) {
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

		n.log.Debug("Obtained content for block", "block", blkHash, "elapsed", time.Since(t0))

		rd := bytes.NewReader(resp)
		var height int64
		if err := binary.Read(rd, binary.LittleEndian, &height); err != nil {
			n.log.Info("failed to read block height in the block response", "error", err)
			continue
		}

		ciBts, err := ktypes.ReadBytes(rd)
		if err != nil {
			n.log.Info("failed to read commit info in the block response", "error", err)
			continue
		}

		var ci ktypes.CommitInfo
		if err = ci.UnmarshalBinary(ciBts); err != nil {
			n.log.Info("failed to unmarshal commit info", "error", err)
			continue
		}

		rawBlk, err := ktypes.ReadBytes(rd)
		if err != nil {
			n.log.Info("failed to read block in the block response", "error", err)
			continue
		}

		return height, rawBlk, &ci, nil
	}
	return 0, nil, nil, ErrBlkNotFound
}

func (n *Node) getBlkHeight(ctx context.Context, height int64) (types.Hash, []byte, *ktypes.CommitInfo, error) {
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

		rd := bytes.NewReader(resp)
		var hash types.Hash

		if _, err := io.ReadFull(rd, hash[:]); err != nil {
			n.log.Warn("failed to read block hash in the block response", "error", err)
			continue
		}

		ciBts, err := ktypes.ReadBytes(rd)
		if err != nil {
			n.log.Info("failed to read commit info in the block response", "error", err)
			continue
		}

		var ci ktypes.CommitInfo
		if err = ci.UnmarshalBinary(ciBts); err != nil {
			n.log.Warn("failed to unmarshal commit info", "error", err)
			continue
		}

		rawBlk, err := ktypes.ReadBytes(rd)
		if err != nil {
			n.log.Warn("failed to read block in the block response", "error", err)
		}

		return hash, rawBlk, &ci, nil
	}
	return types.Hash{}, nil, nil, ErrBlkNotFound
}

// BlockByHeight returns the block by height. If height <= 0, the latest block
// will be returned.
func (n *Node) BlockByHeight(height int64) (types.Hash, *ktypes.Block, *ktypes.CommitInfo, error) {
	if height <= 0 { // I think this is correct since block height starts from 1
		height, _, _, _ = n.bki.Best()
	}
	return n.bki.GetByHeight(height)
}

// BlockByHash returns the block by block hash.
func (n *Node) BlockByHash(hash types.Hash) (*ktypes.Block, *ktypes.CommitInfo, error) {
	return n.bki.Get(hash)
}

// BlockResultByHash returns the block result by block hash.
func (n *Node) BlockResultByHash(hash types.Hash) ([]ktypes.TxResult, error) {
	return n.bki.Results(hash)
}
