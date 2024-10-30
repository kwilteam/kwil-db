package node

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"p2p/node/types"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

type (
	ConsensusReset = types.ConsensusReset
	AckRes         = types.AckRes
)

type ackFrom struct {
	fromPubKey []byte
	res        AckRes
}

type blockProp struct {
	Height    int64
	Hash      types.Hash
	PrevHash  types.Hash
	Stamp     int64
	LeaderSig []byte
	// Replacing *types.Hash
}

func (bp blockProp) String() string {
	return fmt.Sprintf("prop{height:%d hash:%s prevHash:%s}",
		bp.Height, bp.Hash, bp.PrevHash)
}

func (bp blockProp) MarshalBinary() ([]byte, error) {
	// 8 bytes for int64 + 2 hash lengths + 8 bytes for time stamp + len(sig)
	buf := make([]byte, 8+2*types.HashLen+8+len(bp.LeaderSig))
	var c int
	binary.LittleEndian.PutUint64(buf[:8], uint64(bp.Height))
	c += 8
	copy(buf[c:], bp.Hash[:])
	c += types.HashLen
	copy(buf[c:], bp.PrevHash[:])
	c += types.HashLen
	binary.LittleEndian.PutUint64(buf[c:], uint64(bp.Stamp))
	c += 8
	copy(buf[c:], bp.LeaderSig) // c += len(bp.LeaderSig)
	return buf, nil
}

func (bp *blockProp) UnmarshalBinary(data []byte) error {
	const minLeaderSigLen = 0 // don't know length of a valid sig, but it's certainly more than 4 bytes
	if len(data) < 8+2*types.HashLen+8+minLeaderSigLen {
		return fmt.Errorf("insufficient data for blockProp")
	}
	var c int
	bp.Height = int64(binary.LittleEndian.Uint64(data[:8]))
	c += 8
	copy(bp.Hash[:], data[c:c+types.HashLen])
	c += types.HashLen
	copy(bp.PrevHash[:], data[c:])
	c += types.HashLen
	bp.Stamp = int64(binary.LittleEndian.Uint64(data[c:]))
	c += 8
	bp.LeaderSig = data[c:] // c += len(bp.LeaderSig)
	return nil
}

func (bp *blockProp) UnmarshalFromReader(r io.Reader) error {
	// var buf bytes.Buffer
	// buf.Grow(8 + 2*types.HashLen + 8 + 96)
	// _, err := buf.ReadFrom(r)

	// buf := make([]byte, 8+2*types.HashLen+...); io.ReadFull(r, buf)

	// buf, err := io.ReadAll(r)
	buf := make([]byte, 8+2*types.HashLen+8)
	if _, err := io.ReadFull(r, buf); err != nil {
		return fmt.Errorf("reading blockProp: %w", err)
	}
	return bp.UnmarshalBinary(buf)
}

func (n *Node) announceBlkProp(ctx context.Context, blk *types.Block, from peer.ID) {
	rawBlk := types.EncodeBlock(blk)
	blkHash := blk.Hash()
	height := blk.Header.Height

	n.log.Info("announcing proposed block", "hash", blkHash, "height", height,
		"txs", len(blk.Txns), "size", len(rawBlk))

	peers := n.peers()
	if len(peers) == 0 {
		n.log.Warnf("no peers to advertise block to")
		return
	}

	for _, peerID := range peers {
		if peerID == from {
			continue
		}
		prop := blockProp{Height: height, Hash: blkHash, PrevHash: blk.Header.PrevHash,
			Stamp: blk.Header.Timestamp.UnixMilli(), LeaderSig: blk.Signature}
		n.log.Infof("advertising block proposal %s (height %d / txs %d) to peer %v", blkHash, height, len(rawBlk), peerID)
		// resID := annPropMsgPrefix + strconv.Itoa(int(height)) + ":" + prevHash + ":" + blkid
		propID, _ := prop.MarshalBinary()
		err := n.advertiseToPeer(ctx, peerID, ProtocolIDBlockPropose, contentAnn{prop.String(), propID, rawBlk},
			blkSendTimeout)
		if err != nil {
			n.log.Infof(err.Error())
			continue
		}
	}
}

// blkPropStreamHandler is the stream handler for the ProtocolIDBlockPropose
// protocol i.e. proposed block announcements, which originate from the leader,
// but may be re-announced by other validators.
//
// This stream should:
//  1. provide the announcement to the consensus engine (CE)
//  2. if the CE rejects the ann, close stream
//  3. if the CE is ready for this proposed block, request the block
//  4. provide the block contents to the CE
//  5. close the stream
//
// Note that CE decides what to do. For instance, after we provide the full
// block contents, the CE will likely begin executing the blocks. When it is
// done, it will send an ACK/NACK with the
func (n *Node) blkPropStreamHandler(s network.Stream) {
	defer s.Close()

	if n.leader.Load() {
		return
	}

	var prop blockProp
	err := prop.UnmarshalFromReader(s)
	if err != nil {
		n.log.Infof("invalid block proposal message: %v", err)
		return
	}

	height := prop.Height

	if !n.ce.AcceptProposal(height, prop.Hash, prop.PrevHash) {
		// NOTE: if this is ahead of our last commit height, we have to try to catch up
		n.log.Infof("don't want proposal content", height, prop.PrevHash)
		return
	}

	_, err = s.Write([]byte(getMsg))
	if err != nil {
		n.log.Infof("failed to request block proposal contents:", err)
		return
	}

	rd := bufio.NewReader(s)
	blkProp, err := io.ReadAll(rd)
	if err != nil {
		n.log.Infof("failed to read block proposal contents:", err)
		return
	}

	// Q: header first, or full serialized block?

	blk, err := types.DecodeBlock(blkProp)
	if err != nil {
		n.log.Infof("decodeBlock failed for proposal at height %d: %v", height, err)
		return
	}
	if blk.Header.Height != height {
		n.log.Infof("unexpected height: wanted %d, got %d", height, blk.Header.Height)
		return
	}

	annHash := prop.Hash
	hash := blk.Header.Hash()
	if hash != annHash {
		n.log.Infof("unexpected hash: wanted %s, got %s", hash, annHash)
		return
	}

	n.log.Info("processing block proposal", "height", height, "hash", hash)

	go n.ce.NotifyBlockProposal(blk)

	return
}

// sendACK is a callback for the result of validator block execution/precommit.
// After then consensus engine executes the block, this is used to gossip the
// result back to the leader.
func (n *Node) sendACK(ack bool, height int64, blkID types.Hash, appHash *types.Hash) error {
	// fmt.Println("sending ACK", height, ack, blkID, appHash)
	n.ackChan <- types.AckRes{
		ACK:     ack,
		AppHash: appHash,
		BlkHash: blkID,
		Height:  height,
	}
	return nil // actually gossip the nack
}

const (
	TopicACKs = "acks"
)

func (n *Node) startAckGossip(ctx context.Context, ps *pubsub.PubSub) error {
	topicAck, subAck, err := subTopic(ctx, ps, TopicACKs)
	if err != nil {
		return err
	}

	subCanceled := make(chan struct{})

	n.wg.Add(1)
	go func() {
		defer func() {
			<-subCanceled
			topicAck.Close()
			n.wg.Done()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case ack := <-n.ackChan:
				// fmt.Println("received ACK:::", ack.ACK, ack.Height, ack.BlkHash, ack.AppHash)
				ackMsg, _ := ack.MarshalBinary()
				err := topicAck.Publish(ctx, ackMsg)
				if err != nil {
					fmt.Println("Publish:", err)
					// TODO: queue the ack for retry (send back to ackChan or another delayed send queue)
					return
				}
			}

		}
	}()

	me := n.host.ID()

	go func() {
		defer close(subCanceled)
		defer subAck.Cancel()
		for {
			ackMsg, err := subAck.Next(ctx)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					n.log.Infof("subTx.Next:", err)
				}
				return
			}

			// We're only interested if we are the leader.
			if !n.leader.Load() {
				continue // discard, we are just relaying to leader
			}

			if peer.ID(ackMsg.From) == me {
				// n.log.Infof("message from me ignored")
				continue
			}

			var ack AckRes
			err = ack.UnmarshalBinary(ackMsg.Data)
			if err != nil {
				n.log.Infof("failed to decode ACK msg: %v", err)
				continue
			}
			fromPeerID := ackMsg.GetFrom()

			n.log.Infof("received ACK msg from %s (rcvd from %s), data = %x",
				fromPeerID.String(), ackMsg.ReceivedFrom.String(), ackMsg.Message.Data)

			peerPubKey, err := fromPeerID.ExtractPublicKey()
			if err != nil {
				n.log.Infof("failed to extract pubkey from peer ID %v: %v", fromPeerID, err)
				continue
			}
			pubkeyBytes, _ := peerPubKey.Raw() // does not error for secp256k1 or ed25519
			go n.ce.NotifyACK(pubkeyBytes, ack)
		}
	}()

	return nil
}

/* commented because we're probably going with gossipsub
func (n *Node) blkAckStreamHandler(s network.Stream) {
	defer s.Close()

	if !n.leader.Load() {
		return
	}

	// "ack:blkid:appHash" // empty appHash means NACK
	ackMsg := make([]byte, 128)
	nr, err := s.Read(ackMsg)
	if err != nil {
		n.log.Infof("failed to read block proposal ID:", err)
		return
	}
	blkAck, ok := bytes.CutPrefix(ackMsg[:nr], []byte(ackMsg))
	if !ok {
		n.log.Infof("bad block proposal ID:", ackMsg)
		return
	}
	blkID, appHashStr, ok := strings.Cut(string(blkAck), ":")
	if !ok {
		n.log.Infof("bad block proposal ID:", blkAck)
		return
	}

	blkHash, err := types.NewHashFromString(blkID)
	if err != nil {
		n.log.Infof("bad block ID in ack msg:", err)
		return
	}
	isNACK := len(appHashStr) == 0
	if isNACK {
		// do somethign
		n.log.Infof("got nACK for block %v", blkHash)
		return
	}

	appHash, err := types.NewHashFromString(appHashStr)
	if err != nil {
		n.log.Infof("bad block ID in ack msg:", err)
		return
	}

	// as leader, we tally the responses
	n.log.Infof("got ACK for block %v, app hash %v", blkHash, appHash)

	return
}
*/

func (n *Node) sendReset(height int64) error {
	n.resetMsg <- types.ConsensusReset{
		ToHeight: height,
	} // ?
	return nil
}

func subConsensusReset(ctx context.Context, ps *pubsub.PubSub) (*pubsub.Topic, *pubsub.Subscription, error) {
	return subTopic(ctx, ps, TopicReset)
}

func (n *Node) startConsensusResetGossip(ctx context.Context, ps *pubsub.PubSub) error {
	topicReset, subReset, err := subConsensusReset(ctx, ps)
	if err != nil {
		return err
	}

	subCanceled := make(chan struct{})

	n.wg.Add(1)
	go func() {
		defer func() {
			<-subCanceled
			topicReset.Close()
			n.wg.Done()
		}()
		for {
			var resetMsg ConsensusReset
			select {
			case <-ctx.Done():
				return
			case resetMsg = <-n.resetMsg:
			}

			err := topicReset.Publish(ctx, resetMsg.Bytes())
			if err != nil {
				return
			}
		}
	}()

	me := n.host.ID()

	go func() {
		defer close(subCanceled)
		defer subReset.Cancel()
		for {
			resetMsg, err := subReset.Next(ctx)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					n.log.Errorf("Stopping Consensus Reset gossip!", "error", err)
				}
				return
			}

			if string(resetMsg.From) == string(me) {
				continue
			}

			var reset ConsensusReset
			err = reset.UnmarshalBinary(resetMsg.Data)
			if err != nil {
				n.log.Errorf("unable to unmarshal reset msg: %v", err)
				continue
			}

			fromPeerID := resetMsg.GetFrom()

			n.log.Infof("received Consensus Reset msg from %s (rcvd from %s), data = %x",
				fromPeerID, resetMsg.ReceivedFrom, resetMsg.Message.Data)

			// resetSenderPubKey, err := fromPeerID.ExtractPublicKey()
			// if err != nil {
			// 	n.log.Infof("failed to extract pubkey from peer ID %v: %v", fromPeerID, err)
			// 	continue
			// }
			// go n.ce.NotifyReset(reset, resetSenderPubKey)
		}
	}()

	return nil
}
