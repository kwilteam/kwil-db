package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/core/crypto"
)

type ConsensusReset struct {
	ToHeight int64
	TxIDs    []Hash
}

func (cr ConsensusReset) String() string {
	return fmt.Sprintf("ConsensusReset{Height: %d}", cr.ToHeight)
}

func (cr ConsensusReset) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint64(cr.ToHeight))
	binary.Write(buf, binary.LittleEndian, uint64(len(cr.TxIDs)))
	for _, txID := range cr.TxIDs {
		buf.Write(txID[:])
	}

	return buf.Bytes()
}

func (cr ConsensusReset) MarshalBinary() ([]byte, error) {
	return cr.Bytes(), nil
}

func (cr *ConsensusReset) UnmarshalBinary(data []byte) error {
	if len(data) < 16 {
		return errors.New("invalid ConsensusReset data")
	}

	buf := bytes.NewBuffer(data)

	var height uint64
	if err := binary.Read(buf, binary.LittleEndian, &height); err != nil {
		return err
	}
	cr.ToHeight = int64(height)

	var numTxIDs uint64
	if err := binary.Read(buf, binary.LittleEndian, &numTxIDs); err != nil {
		return err
	}
	cr.TxIDs = make([]Hash, numTxIDs)

	for i := range cr.TxIDs {
		if _, err := buf.Read(cr.TxIDs[i][:]); err != nil {
			return err
		}
	}

	return nil
}

type AckRes struct {
	ACK     bool
	Height  int64
	BlkHash Hash
	AppHash *Hash

	// Signature
	PubKeyType crypto.KeyType
	PubKey     []byte
	Signature  []byte
}

func (ar AckRes) ack() string {
	if ar.ACK {
		return "ACK"
	}
	return "nACK"
}

func (ar AckRes) String() string {
	if ar.ACK {
		return fmt.Sprintf("%s: height: %d, block %v, appHash %v", ar.ack(), ar.Height, ar.BlkHash, ar.AppHash)
	}
	return ar.ack()
}

func (ar AckRes) MarshalBinary() ([]byte, error) {
	if ar.ACK && ar.AppHash == nil {
		return nil, errors.New("app hash is required for ACK")
	}

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, ar.ACK); err != nil {
		return nil, fmt.Errorf("failed to write ACK: %v", err)
	}

	if err := binary.Write(&buf, binary.LittleEndian, uint64(ar.Height)); err != nil {
		return nil, fmt.Errorf("failed to write height in AckRes: %v", err)
	}

	if err := binary.Write(&buf, binary.LittleEndian, ar.BlkHash[:]); err != nil {
		return nil, fmt.Errorf("failed to write block hash in AckRes: %v", err)
	}

	if ar.AppHash != nil {
		if err := binary.Write(&buf, binary.LittleEndian, true); err != nil {
			return nil, fmt.Errorf("failed to write app hash flag in AckRes: %v", err)
		}

		if err := binary.Write(&buf, binary.LittleEndian, ar.AppHash[:]); err != nil {
			return nil, fmt.Errorf("failed to write app hash in AckRes: %v", err)
		}
	} else {
		if err := binary.Write(&buf, binary.LittleEndian, false); err != nil {
			return nil, fmt.Errorf("failed to write app hash flag in AckRes: %v", err)
		}
	}

	if err := binary.Write(&buf, binary.LittleEndian, ar.PubKeyType); err != nil {
		return nil, fmt.Errorf("failed to write key type in AckRes: %v", err)
	}

	if err := binary.Write(&buf, binary.LittleEndian, uint64(len(ar.PubKey))); err != nil {
		return nil, fmt.Errorf("failed to write key length in AckRes: %v", err)
	}

	if _, err := buf.Write(ar.PubKey); err != nil {
		return nil, fmt.Errorf("failed to write key in AckRes: %v", err)
	}

	if err := binary.Write(&buf, binary.LittleEndian, uint64(len(ar.Signature))); err != nil {
		return nil, fmt.Errorf("failed to write signature length in AckRes: %v", err)
	}

	if _, err := buf.Write(ar.Signature); err != nil {
		return nil, fmt.Errorf("failed to write signature in AckRes: %v", err)
	}

	return buf.Bytes(), nil
}

func (ar *AckRes) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)

	if err := binary.Read(buf, binary.LittleEndian, &ar.ACK); err != nil {
		return fmt.Errorf("failed to read ACK: %v", err)
	}

	var height uint64
	if err := binary.Read(buf, binary.LittleEndian, &height); err != nil {
		return fmt.Errorf("failed to read height in AckRes: %v", err)
	}
	ar.Height = int64(height)

	if _, err := buf.Read(ar.BlkHash[:]); err != nil {
		return fmt.Errorf("failed to read block hash in AckRes: %v", err)
	}

	var hasAppHash bool
	if err := binary.Read(buf, binary.LittleEndian, &hasAppHash); err != nil {
		return fmt.Errorf("failed to read app hash flag in AckRes: %v", err)
	}

	if hasAppHash {
		var appHash Hash
		if _, err := buf.Read(appHash[:]); err != nil {
			return fmt.Errorf("failed to read app hash in AckRes: %v", err)
		}
		ar.AppHash = &appHash
	}

	if err := binary.Read(buf, binary.LittleEndian, &ar.PubKeyType); err != nil {
		return fmt.Errorf("failed to read key type in AckRes: %v", err)
	}

	var keyLen uint64
	if err := binary.Read(buf, binary.LittleEndian, &keyLen); err != nil {
		return fmt.Errorf("failed to read key length in AckRes: %v", err)
	}
	if keyLen > 0 {
		ar.PubKey = make([]byte, keyLen)
		if _, err := buf.Read(ar.PubKey); err != nil {
			return fmt.Errorf("failed to read key in AckRes: %v", err)
		}
	}

	var sigLen uint64
	if err := binary.Read(buf, binary.LittleEndian, &sigLen); err != nil {
		return fmt.Errorf("failed to read signature length in AckRes: %v", err)
	}

	if sigLen > 0 {
		ar.Signature = make([]byte, sigLen)
		if _, err := buf.Read(ar.Signature); err != nil {
			return fmt.Errorf("failed to read signature in AckRes: %v", err)
		}
	}

	return nil
}

type DiscoveryRequest struct{}

func (dr DiscoveryRequest) String() string {
	return "DiscoveryRequest"
}

type DiscoveryResponse struct {
	BestHeight int64
}

func (dr DiscoveryResponse) String() string {
	return fmt.Sprintf("DiscoveryMsg{BestHeight: %d}", dr.BestHeight)
}

func (dr DiscoveryResponse) Bytes() []byte {
	return binary.LittleEndian.AppendUint64(nil, uint64(dr.BestHeight))
}

func (dr DiscoveryResponse) MarshalBinary() ([]byte, error) {
	return dr.Bytes(), nil
}

func (dr *DiscoveryResponse) UnmarshalBinary(data []byte) error {
	if len(data) != 8 {
		return errors.New("invalid DiscoveryMsg data")
	}
	dr.BestHeight = int64(binary.LittleEndian.Uint64(data))
	return nil
}
