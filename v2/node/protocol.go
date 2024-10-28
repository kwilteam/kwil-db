package node

import (
	"bytes"
	"context"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"p2p/node/types"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	ProtocolIDDiscover    protocol.ID = "/kwil/discovery/1.0.0"
	ProtocolIDTx          protocol.ID = "/kwil/tx/1.0.0"
	ProtocolIDTxAnn       protocol.ID = "/kwil/txann/1.0.0"
	ProtocolIDBlockHeight protocol.ID = "/kwil/blkheight/1.0.0"
	ProtocolIDBlock       protocol.ID = "/kwil/blk/1.0.0"
	ProtocolIDBlkAnn      protocol.ID = "/kwil/blkann/1.0.0"
	// ProtocolIDBlockHeader protocol.ID = "/kwil/blkhdr/1.0.0"

	ProtocolIDBlockPropose protocol.ID = "/kwil/blkprop/1.0.0"
	// ProtocolIDACKProposal  protocol.ID = "/kwil/blkack/1.0.0"

	getMsg = "get" // context dependent, in open stream convo
)

func requestFrom(ctx context.Context, host host.Host, peer peer.ID, resID []byte,
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

	return request(txStream, resID, readLimit)
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

var noData = []byte{0}

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

// blockInitMsg is for ProtocolIDBlkAnn "/kwil/blkann/1.0.0"
type blockInitMsg struct {
	Hash    types.Hash
	Height  int64
	AppHash types.Hash // could be in the content/response
}

var _ encoding.BinaryMarshaler = blockInitMsg{}
var _ encoding.BinaryMarshaler = (*blockInitMsg)(nil)

func (m blockInitMsg) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	_, err := m.WriteTo(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ encoding.BinaryUnmarshaler = (*blockInitMsg)(nil)

func (m *blockInitMsg) UnmarshalBinary(data []byte) error {
	_, err := m.ReadFrom(bytes.NewReader(data))
	return err
}

var _ io.WriterTo = (*blockInitMsg)(nil)

func (m *blockInitMsg) WriteTo(w io.Writer) (int64, error) {
	var n int
	nw, err := w.Write(m.Hash[:])
	if err != nil {
		return int64(nw), err
	}
	n += nw

	hBts := binary.LittleEndian.AppendUint64(nil, uint64(m.Height))
	nw, err = w.Write(hBts)
	if err != nil {
		return int64(n), err
	}
	n += nw

	nw, err = w.Write(m.AppHash[:])
	if err != nil {
		return int64(n), err
	}
	n += nw

	return int64(n), nil
}

var _ io.ReaderFrom = (*blockInitMsg)(nil)

func (m *blockInitMsg) ReadFrom(r io.Reader) (int64, error) {
	nr, err := io.ReadFull(r, m.Hash[:])
	if err != nil {
		return int64(nr), err
	}
	n := int64(nr)
	if err := binary.Read(r, binary.LittleEndian, &m.Height); err != nil {
		return n, err
	}
	n += 8
	if nr, err := io.ReadFull(r, m.AppHash[:]); err != nil {
		return n + int64(nr), err
	}
	n += int64(nr)
	return n, nil
}

// blockHeightReq is for ProtocolIDBlockHeight "/kwil/blkheight/1.0.0"
type blockHeightReq struct {
	Height int64
}

var _ encoding.BinaryMarshaler = blockHeightReq{}
var _ encoding.BinaryMarshaler = (*blockHeightReq)(nil)

func (r blockHeightReq) MarshalBinary() ([]byte, error) {
	return binary.LittleEndian.AppendUint64(nil, uint64(r.Height)), nil
}

func (r *blockHeightReq) UnmarshalBinary(data []byte) error {
	if len(data) != 8 {
		return errors.New("unexpected data length")
	}
	r.Height = int64(binary.LittleEndian.Uint64(data))
	return nil
}

var _ io.WriterTo = (*blockHeightReq)(nil)

func (r blockHeightReq) WriteTo(w io.Writer) (int64, error) {
	bts, _ := r.MarshalBinary()
	n, err := w.Write(bts)
	return int64(n), err
}

var _ io.ReaderFrom = (*blockHeightReq)(nil)

func (r *blockHeightReq) ReadFrom(rd io.Reader) (int64, error) {
	hBts := make([]byte, 8)
	n, err := io.ReadFull(rd, hBts)
	if err != nil {
		return int64(n), err
	}
	r.Height = int64(binary.LittleEndian.Uint64(hBts))
	return int64(n), err
}

// blockHashReq is for ProtocolIDBlock "/kwil/blk/1.0.0"
type blockHashReq struct {
	Hash types.Hash
}

var _ encoding.BinaryMarshaler = blockHashReq{}
var _ encoding.BinaryMarshaler = (*blockHashReq)(nil)

func (r blockHashReq) MarshalBinary() ([]byte, error) {
	return r.Hash[:], nil
}

func (r *blockHashReq) UnmarshalBinary(data []byte) error {
	if len(data) != types.HashLen {
		return fmt.Errorf("invalid hash length")
	}
	copy(r.Hash[:], data)
	return nil
}

var _ io.WriterTo = (*blockHashReq)(nil)

func (r blockHashReq) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(r.Hash[:])
	return int64(n), err
}

var _ io.ReaderFrom = (*blockHashReq)(nil)

func (r *blockHashReq) ReadFrom(rd io.Reader) (int64, error) {
	n, err := io.ReadFull(rd, r.Hash[:])
	return int64(n), err
}

// txHashReq is for ProtocolIDTx "/kwil/tx/1.0.0"
type txHashReq struct {
	blockHashReq // just embed the methods for the identical block hash request for now
}

func newTxHashReq(hash types.Hash) txHashReq {
	return txHashReq{blockHashReq{Hash: hash}}
}

// txHashAnn is for ProtocolIDTxAnn "/kwil/txann/1.0.0"
type txHashAnn struct {
	blockHashReq
}

func newTxHashAnn(hash types.Hash) txHashAnn {
	return txHashAnn{blockHashReq{Hash: hash}}
}
