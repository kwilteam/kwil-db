package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
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

// In scenarios where the leader is trying to catchup, there is a possibility
// that the leader syncs to a height which is far behind the network's best height,
// and leader starts proposing the blocks from that height. In such cases, the
// Validators upon hearing a new block proposal for already committed block should
// respond to the leader with a Nack, providing leader feedback about it's status
// including the blk proposal of the height it is at, with the leader's signature.
// Leader can use this feedback to eventually catch up with the network.

// NackStatus desribes the reason for a nack response.
type NackStatus string

const (
	// If the block validation fails either due to invalid header info such as
	// AppHash or the ValidatorHash or Invalid Merkle hash etc.
	NackStatusInvalidBlock NackStatus = "invalid_block"
	// If leader proposes a new block for an already committed height, indicating
	// that the leader may potentially be out of sync with the rest of the network.
	// This requires the validator to prove to the leader that the block is indeed
	// committed by sending the block header with the leaders signature in the Vote.
	NackStatusOutOfSync NackStatus = "out_of_sync"
	// other unknown miscellaneous reasons for nack
	NackStatusUnknown NackStatus = "unknown"
)

func (ns NackStatus) String() string {
	return string(ns)
}

type OutOfSyncProof struct {
	Header    *types.BlockHeader
	Signature []byte
}
type AckRes struct {
	ACK bool
	// only required if ACK is false
	NackStatus *NackStatus
	Height     int64
	BlkHash    Hash
	// only required if ACK is true
	AppHash *Hash
	// optional, only required if the nack status is NackStatusOutOfSync
	OutOfSyncProof *OutOfSyncProof

	// Signature
	Signature *Signature
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
	} else if !ar.ACK {
		if ar.AppHash != nil {
			return nil, errors.New("app hash is not allowed for nACK")
		} else if ar.NackStatus == nil {
			return nil, errors.New("nack status is required for nACK")
		} else if *ar.NackStatus == NackStatusOutOfSync && ar.OutOfSyncProof == nil {
			return nil, errors.New("proof is required for out of sync nack")
		}
	}
	if ar.Signature == nil {
		return nil, errors.New("signature is required in the AckRes")
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

	if ar.ACK {
		// app hash
		if err := binary.Write(&buf, binary.LittleEndian, ar.AppHash[:]); err != nil {
			return nil, fmt.Errorf("failed to write app hash in AckRes: %v", err)
		}
	} else {
		// nack status
		if err := types.WriteString(&buf, (*ar.NackStatus).String()); err != nil {
			return nil, fmt.Errorf("failed to write nack status in AckRes: %v", err)
		}
		// if nack status is NackStatusOutOfSync, write out of sync proof
		if *ar.NackStatus == NackStatusOutOfSync {
			// write header
			headerBts := types.EncodeBlockHeader(ar.OutOfSyncProof.Header)
			if err := types.WriteBytes(&buf, headerBts); err != nil {
				return nil, fmt.Errorf("failed to write header in AckRes: %v", err)
			}
			// write signature
			if err := types.WriteBytes(&buf, ar.OutOfSyncProof.Signature); err != nil {
				return nil, fmt.Errorf("failed to write signature in AckRes: %v", err)
			}
		}
	}

	sigBts := EncodeSignature(ar.Signature)
	if err := types.WriteBytes(&buf, sigBts); err != nil {
		return nil, fmt.Errorf("failed to write signature in AckRes: %v", err)
	}

	// if err := types.WriteString(&buf, string(ar.PubKeyType)); err != nil {
	// 	return nil, fmt.Errorf("failed to write key type in AckRes: %v", err)
	// }

	// if err := types.WriteBytes(&buf, ar.PubKey); err != nil {
	// 	return nil, fmt.Errorf("failed to write key in AckRes: %v", err)
	// }

	// if err := types.WriteBytes(&buf, ar.Signature); err != nil {
	// 	return nil, fmt.Errorf("failed to write signature in AckRes: %v", err)
	// }

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

	if ar.ACK {
		// Read app hash
		var appHash Hash
		if _, err := buf.Read(appHash[:]); err != nil {
			return fmt.Errorf("failed to read app hash in AckRes: %v", err)
		}
		ar.AppHash = &appHash
	} else {
		// Read nack status
		ns, err := types.ReadString(buf)
		if err != nil {
			return fmt.Errorf("failed to read nack status in AckRes: %v", err)
		}
		nackStatus := NackStatus(ns)
		ar.NackStatus = &nackStatus

		// if nack status is NackStatusOutOfSync, read out of sync proof
		if *ar.NackStatus == NackStatusOutOfSync {
			headerBts, err := types.ReadBytes(buf)
			if err != nil {
				return fmt.Errorf("failed to read header in AckRes: %v", err)
			}
			header, err := types.DecodeBlockHeader(bytes.NewBuffer(headerBts))
			if err != nil {
				return fmt.Errorf("failed to decode header in AckRes: %v", err)
			}

			sigBts, err := types.ReadBytes(buf)
			if err != nil {
				return fmt.Errorf("failed to read signature in AckRes: %v", err)
			}

			ar.OutOfSyncProof = &OutOfSyncProof{
				Header:    header,
				Signature: sigBts,
			}
		}
	}

	sigBts, err := types.ReadBytes(buf)
	if err != nil {
		return fmt.Errorf("failed to read signature in AckRes: %v", err)
	}
	sig, err := DecodeSignature(sigBts)
	if err != nil {
		return fmt.Errorf("failed to decode signature in AckRes: %v", err)
	}
	ar.Signature = sig

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
