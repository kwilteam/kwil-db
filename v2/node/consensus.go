package node

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"p2p/node/types"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

type AckRes = types.AckRes

type ackFrom struct {
	fromPubKey []byte
	res        AckRes
}

type blockProp struct {
	Height   int64
	Hash     types.Hash
	PrevHash types.Hash
}

func (bp blockProp) String() string {
	return fmt.Sprintf("prop{height:%d hash:%s prevHash:%s}",
		bp.Height, bp.Hash, bp.PrevHash)
}

func (bp blockProp) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 8+2*types.HashLen) // 8 bytes for int64 + 2 hash lengths
	binary.LittleEndian.PutUint64(buf[:8], uint64(bp.Height))
	copy(buf[8:], bp.Hash[:])
	copy(buf[8+types.HashLen:], bp.PrevHash[:])
	return buf, nil
}

func (bp *blockProp) UnmarshalBinary(data []byte) error {
	if len(data) < 8+2*types.HashLen {
		return fmt.Errorf("insufficient data for blockProp")
	}
	bp.Height = int64(binary.LittleEndian.Uint64(data[:8]))
	copy(bp.Hash[:], data[8:8+types.HashLen])
	copy(bp.PrevHash[:], data[8+types.HashLen:])
	return nil
}

func (bp *blockProp) UnmarshalFromReader(r io.Reader) error {
	buf := make([]byte, 8+2*types.HashLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return fmt.Errorf("reading blockProp: %w", err)
	}
	return bp.UnmarshalBinary(buf)
}

func (n *Node) announceBlkProp(ctx context.Context, blk *types.Block, from peer.ID) {
	rawBlk := types.EncodeBlock(blk)
	blkHash := blk.Hash()
	height := blk.Header.Height

	log.Printf("ANNOUNCING PROPOSED BLOCK %v / %d size = %d, txs = %d\n",
		blkHash, height, len(rawBlk), len(blk.Txns))

	peers := n.peers()
	if len(peers) == 0 {
		log.Println("no peers to advertise block to")
		return
	}

	for _, peerID := range peers {
		if peerID == from {
			continue
		}
		prop := blockProp{Height: height, Hash: blkHash, PrevHash: blk.Header.PrevHash}
		log.Printf("advertising block proposal %s (height %d / txs %d) to peer %v", blkHash, height, len(rawBlk), peerID)
		// resID := annPropMsgPrefix + strconv.Itoa(int(height)) + ":" + prevHash + ":" + blkid
		propID, _ := prop.MarshalBinary()
		err := advertiseToPeer(ctx, n.host, peerID, ProtocolIDBlockPropose, contentAnn{prop.String(), propID, rawBlk})
		if err != nil {
			log.Println(err)
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
		log.Println("invalid block proposal message:", err)
		return
	}

	height := prop.Height

	if !n.ce.AcceptProposalID(height, prop.PrevHash) {
		// NOTE: if this is ahead of our last commit height, we have to try to catch up
		log.Println("don't want proposal content", height, prop.PrevHash)
		return
	}

	_, err = s.Write([]byte(getMsg))
	if err != nil {
		log.Println("failed to request block proposal contents:", err)
		return
	}

	rd := bufio.NewReader(s)
	blkProp, err := io.ReadAll(rd)
	if err != nil {
		log.Println("failed to read block proposal contents:", err)
		return
	}

	// Q: header first, or full serialized block?

	blk, err := types.DecodeBlock(blkProp)
	if err != nil {
		log.Printf("decodeBlock failed for proposal at height %d: %v", height, err)
		return
	}
	if blk.Header.Height != height {
		log.Printf("unexpected height: wanted %d, got %d", height, blk.Header.Height)
		return
	}

	annHash := prop.Hash
	hash := blk.Header.Hash()
	if hash != annHash {
		log.Printf("unexpected hash: wanted %s, got %s", hash, annHash)
		return
	}

	log.Println("processing prop for", hash)

	go n.ce.ProcessProposal(blk, func(ack bool, appHash *types.Hash) error {
		return n.sendACK(ack, hash, appHash)
	})

	return
}

// sendACK is a callback for the result of validator block execution/precommit.
// After then consensus engine executes the block, this is used to gossip the
// result back to the leader.
func (n *Node) sendACK(ack bool, blkID types.Hash, appHash *types.Hash) error {
	n.ackChan <- types.AckRes{
		ACK:     ack,
		AppHash: appHash,
		BlkHash: blkID,
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
					log.Println("subTx.Next:", err)
				}
				return
			}

			// We're only interested if we are the leader.
			if !n.leader.Load() {
				continue // discard, we are just relaying to leader
			}

			if peer.ID(ackMsg.From) == me {
				// log.Println("message from me ignored")
				continue
			}

			var ack AckRes
			err = ack.UnmarshalBinary(ackMsg.Data)
			if err != nil {
				log.Printf("failed to decode ACK msg: %v", err)
				continue
			}
			fromPeerID := ackMsg.GetFrom()

			log.Printf("received ACK msg from %v (rcvd from %s), data = %x",
				fromPeerID, ackMsg.ReceivedFrom, ackMsg.Message.Data)

			peerPubKey, err := fromPeerID.ExtractPublicKey()
			if err != nil {
				log.Printf("failed to extract pubkey from peer ID %v: %v", fromPeerID, err)
				continue
			}
			pubkeyBytes, _ := peerPubKey.Raw() // does not error for secp256k1 or ed25519
			go n.ce.ProcessACK(pubkeyBytes, ack)
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
		log.Println("failed to read block proposal ID:", err)
		return
	}
	blkAck, ok := bytes.CutPrefix(ackMsg[:nr], []byte(ackMsg))
	if !ok {
		log.Println("bad block proposal ID:", ackMsg)
		return
	}
	blkID, appHashStr, ok := strings.Cut(string(blkAck), ":")
	if !ok {
		log.Println("bad block proposal ID:", blkAck)
		return
	}

	blkHash, err := types.NewHashFromString(blkID)
	if err != nil {
		log.Println("bad block ID in ack msg:", err)
		return
	}
	isNACK := len(appHashStr) == 0
	if isNACK {
		// do somethign
		log.Printf("got nACK for block %v", blkHash)
		return
	}

	appHash, err := types.NewHashFromString(appHashStr)
	if err != nil {
		log.Println("bad block ID in ack msg:", err)
		return
	}

	// as leader, we tally the responses
	log.Printf("got ACK for block %v, app hash %v", blkHash, appHash)

	return
}
*/
