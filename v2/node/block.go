package node

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"p2p/node/types"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	blkReadLimit  = 300_000_000
	blkGetTimeout = 90 * time.Second
)

func (n *Node) blkGetStreamHandler(s network.Stream) {
	defer s.Close()

	req := make([]byte, 128)
	// io.ReadFull(s, req)
	nr, err := s.Read(req)
	if err != nil && err != io.EOF {
		fmt.Println("bad get blk req", err)
		return
	}
	req, ok := bytes.CutPrefix(req[:nr], []byte(getBlkMsgPrefix))
	if !ok {
		fmt.Println("bad get blk request")
		return
	}
	blkid := strings.TrimSpace(string(req))
	log.Printf("requested blkid: %q", blkid)
	blkHash, err := types.NewHashFromString(blkid)
	if err != nil {
		fmt.Println("invalid block ID", err)
		return
	}

	blk, appHash, err := n.bki.Get(blkHash)
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

	req := make([]byte, 128)
	// io.ReadFull(s, req)
	nr, err := s.Read(req)
	if err != nil && err != io.EOF {
		fmt.Println("bad get blk req", err)
		return
	}
	req, ok := bytes.CutPrefix(req[:nr], []byte(getBlkHeightMsgPrefix))
	if !ok {
		fmt.Println("bad get blk(height) request")
		return
	}
	heightStr := strings.TrimSpace(string(req))
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		fmt.Println("invalid block ID", err)
		return
	}
	log.Printf("requested block height: %q", height)

	hash, blk, appHash, err := n.bki.GetByHeight(height)
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

	req := make([]byte, 256)
	nr, err := s.Read(req)
	if err != nil && err != io.EOF {
		log.Println("bad blk ann req", err)
		return
	}
	req, ok := bytes.CutPrefix(req[:nr], []byte(annBlkMsgPrefix))
	if !ok {
		log.Println("bad blk ann request")
		return
	}
	if len(req) <= 64 {
		log.Println("short blk ann request")
		return
	}
	blkid, after, cut := strings.Cut(string(req), ":")
	if !cut {
		log.Println("invalid blk ann request")
		return
	}
	var appHash types.Hash
	heightStr, appHashStr, cut := strings.Cut(after, ":")
	if cut {
		appHash, err = types.NewHashFromString(appHashStr)
		if err != nil {
			log.Println("BAD appHash in blk ann request", err)
			return
		}
	}

	blkHash, err := types.NewHashFromString(blkid)
	if err != nil {
		log.Printf("invalid block id: %v", err)
		return
	}
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		log.Printf("invalid height in blk ann request: %v", err)
		return
	}
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
		gotHeight, rawBlk, appHash, err = n.getBlkWithRetry(ctx, blkid, 500*time.Millisecond, 10)
		if err != nil {
			log.Printf("unable to retrieve tx %v: %v", blkid, err)
			return
		}
		if gotHeight != height {
			log.Printf("getblk response had unexpected height: wanted %d, got %d", height, gotHeight)
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
		var appHashStr string
		if !appHash.IsZero() {
			appHashStr = appHash.String()
		}
		n.announceRawBlk(context.Background(), blkid, height, rawBlk, appHashStr, s.Conn().RemotePeer())
	}()
}

func (n *Node) announceBlk(ctx context.Context, blk *types.Block, appHash types.Hash, from peer.ID) {
	fmt.Println("announceBlk", blk.Header.Height, blk.Header.Hash(), appHash, from)
	rawBlk := types.EncodeBlock(blk)
	var appHashStr string
	if !appHash.IsZero() {
		appHashStr = appHash.String()
	}
	n.announceRawBlk(ctx, blk.Hash().String(), blk.Header.Height, rawBlk, appHashStr, from)
	return
}

func (n *Node) announceRawBlk(ctx context.Context, blkid string, height int64,
	rawBlk []byte, appHash string, from peer.ID) {
	peers := n.peers()
	if len(peers) == 0 {
		log.Println("no peers to advertise block to")
		return
	}

	for _, peerID := range peers {
		if peerID == from {
			continue
		}
		log.Printf("advertising block %s (height %d / txs %d) to peer %v", blkid, height, len(rawBlk), peerID)
		resID := annBlkMsgPrefix + blkid + ":" + strconv.Itoa(int(height))
		if appHash != "" {
			resID += ":" + appHash
		}
		err := advertiseToPeer(ctx, n.host, peerID, ProtocolIDBlkAnn, contentAnn{resID, []byte(resID), rawBlk})
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func (n *Node) getBlkWithRetry(ctx context.Context, blkid string, baseDelay time.Duration,
	maxAttempts int) (int64, []byte, types.Hash, error) {
	var attempts int
	for {
		height, raw, appHash, err := n.getBlk(ctx, blkid)
		if err == nil {
			return height, raw, appHash, nil
		}

		log.Printf("unable to retrieve block %v (%v), waiting to retry", blkid, err)

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

func (n *Node) getBlk(ctx context.Context, blkid string) (int64, []byte, types.Hash, error) {
	for _, peer := range n.peers() {
		log.Printf("requesting block %v from %v", blkid, peer)
		t0 := time.Now()
		resID := getBlkMsgPrefix + blkid
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

		log.Printf("obtained content for block %q in %v", blkid, time.Since(t0))

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
		resID := getBlkHeightMsgPrefix + strconv.FormatInt(height, 10)
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

		log.Printf("obtained content for block number %q in %v", height, time.Since(t0))

		var hash, appHash types.Hash
		copy(hash[:], resp[:types.HashLen])
		copy(appHash[:], resp[types.HashLen:types.HashLen*2])
		rawBlk := resp[types.HashLen*2:]

		return hash, appHash, rawBlk, nil
	}
	return types.Hash{}, types.ZeroHash, nil, ErrBlkNotFound
}
