package node

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	ProtocolIDDiscover protocol.ID = "/kwil/discovery/1.0.0"
	ProtocolIDTx       protocol.ID = "/kwil/tx/1.0.0"
	ProtocolIDTxAnn    protocol.ID = "/kwil/txann/1.0.0"
	ProtocolIDBlock    protocol.ID = "/kwil/blk/1.0.0"
	ProtocolIDBlkAnn   protocol.ID = "/kwil/blkann/1.0.0"

	ProtocolIDBlockPropose protocol.ID = "/kwil/blkprop/1.0.0"
	// ProtocolIDACKProposal  protocol.ID = "/kwil/blkack/1.0.0"

	// These prefixes are protocol specific. They are intended to future proof
	// the protocol handlers so different proto versions can be handled with
	// shared code.
	annTxMsgPrefix  = "txann:"
	getTxMsgPrefix  = "gettx:"
	annBlkMsgPrefix = "blkann:"
	getBlkMsgPrefix = "getblk:"

	annPropMsgPrefix = "prop:"
	annAckMsgPrefix  = "ack:"

	getMsg = "get" // context dependent, in open stream convo
)

func requestFrom(ctx context.Context, host host.Host, peer peer.ID, resID string,
	proto protocol.ID, readLimit int64) ([]byte, error) {
	txStream, err := host.NewStream(ctx, peer, proto)
	if err != nil {
		return nil, err
	}
	defer txStream.Close()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(txGetTimeout)
	}

	txStream.SetDeadline(deadline)

	return request(txStream, []byte(resID), readLimit)
}

func request(rw io.ReadWriter, reqMsg []byte, readLimit int64) ([]byte, error) {
	_, err := rw.Write(reqMsg)
	if err != nil {
		return nil, fmt.Errorf("resource get request failed: %v", err)
	}

	rawTx, err := readResp(rw, readLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource get response: %v", err)
	}
	return rawTx, nil
}

var noData = []byte("0")

func readResp(rd io.Reader, limit int64) ([]byte, error) {
	rd = io.LimitReader(rd, limit)
	resp, err := io.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, ErrNoResponse
	}
	if bytes.Equal(resp, noData) {
		return nil, ErrNotFound
	}
	return resp, nil
}

type contentAnn struct {
	cType   string
	ann     []byte // may be cType if self-describing
	content []byte
}

func (ca contentAnn) String() string {
	return ca.cType
}

// advertiseToPeer sends a lightweight advertisement to a connected peer.
// The stream remains open in case the peer wants to request the content .
func advertiseToPeer(ctx context.Context, host host.Host, peerID peer.ID, proto protocol.ID, ann contentAnn) error {
	s, err := host.NewStream(ctx, peerID, proto)
	if err != nil {
		return fmt.Errorf("failed to open stream to peer: %w", err)
	}

	// Send a lightweight advertisement with the object ID
	_, err = s.Write([]byte(ann.ann))
	if err != nil {
		return fmt.Errorf("send content ID failed: %w", err)
	}

	log.Printf("advertised content %s to peer %s", ann, peerID)

	// Keep the stream open for potential content requests
	go func() {
		defer s.Close()

		req := make([]byte, 128)
		n, err := s.Read(req)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Println("bad get blk req", err)
			return
		}
		if n == 0 { // they didn't want it
			return
		}
		req = req[:n]
		req, ok := bytes.CutPrefix(req, []byte(getMsg))
		if !ok {
			log.Printf("bad get request for %s: %v", ann, req)
			return
		}
		s.Write(ann.content)
	}()

	return nil
}
