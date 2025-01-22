package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils"
)

// CommitInfo includes the information about the commit of the block.
// Such as the signatures of the validators aggreeing to the block.
type CommitInfo struct {
	AppHash      Hash
	Votes        []*VoteInfo
	ParamUpdates types.ParamUpdates
}

type NodeStatus struct {
	Role            string                   `json:"role"`
	CatchingUp      bool                     `json:"catching_up"`
	CommittedHeader *types.BlockHeader       `json:"committed_header"`
	CommitInfo      *CommitInfo              `json:"commit_info"`
	Params          *types.NetworkParameters `json:"params"`
}

type AckStatus int

const (
	AckStatusDisagree AckStatus = iota
	AckStatusAgree
	AckStatusDiverge
)

func (ack *AckStatus) String() string {
	switch *ack {
	case AckStatusDisagree:
		return "disagree"
	case AckStatusAgree:
		return "agree"
	case AckStatusDiverge:
		return "diverge"
	default:
		return "unknown"
	}
}

type VoteInfo struct {
	// VoteSignature is the signature of the blkHash + nack | blkHash + ack + appHash
	Signature Signature

	// Ack is set to true if the validator agrees with the block
	// in terms of the AppHash, ValidatorSet, MerkleRoot of Txs etc.
	AckStatus AckStatus
	// AppHash is optional, it set only if the AckStatus is AckStatusDivereged.
	// AppHash is implied to be the AppHash in the CommitInfo if the AckStatus is AckStatusAgree.
	// AppHash is nil if the AckStatus is AckStatusDisagree.
	AppHash *Hash
}

type Signature struct {
	PubKeyType crypto.KeyType
	PubKey     []byte // public key of the validator

	Data []byte
}

func (s *Signature) WriteTo(w io.Writer) (int64, error) {
	cw := utils.NewCountingWriter(w)
	// PubKeyType
	if _, err := s.PubKeyType.WriteTo(cw); err != nil {
		return cw.Written(), err
	}

	// PubKey Length
	if err := types.WriteBytes(cw, s.PubKey); err != nil {
		return cw.Written(), err
	}

	// Signature Data Length
	if err := types.WriteBytes(cw, s.Data); err != nil {
		return cw.Written(), err
	}

	return cw.Written(), nil
}

func (s *Signature) ReadFrom(r io.Reader) (int64, error) {
	cr := utils.NewCountingReader(r)

	var kt crypto.KeyType
	_, err := kt.ReadFrom(cr)
	if err != nil {
		return cr.ReadCount(), fmt.Errorf("failed to read public key type: %w", err)
	}
	s.PubKeyType = kt

	pubKey, err := types.ReadBytes(cr)
	if err != nil {
		return cr.ReadCount(), fmt.Errorf("failed to read public key: %w", err)
	}
	s.PubKey = pubKey

	sig, err := types.ReadBytes(cr)
	if err != nil {
		return cr.ReadCount(), fmt.Errorf("failed to read signature: %w", err)
	}
	s.Data = sig

	return cr.ReadCount(), nil
}

func (v *VoteInfo) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	_, err := v.Signature.WriteTo(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to write signature: %w", err)
	}

	if err := binary.Write(&buf, binary.LittleEndian, int32(v.AckStatus)); err != nil {
		return nil, fmt.Errorf("failed to write ack status: %w", err)
	}

	if v.AckStatus == AckStatusDiverge {
		if v.AppHash == nil {
			return nil, errors.New("missing app hash for diverged vote")
		}

		if _, err := buf.Write(v.AppHash[:]); err != nil {
			return nil, fmt.Errorf("failed to write app hash: %w", err)
		}
	}

	return buf.Bytes(), nil
}

func (v *VoteInfo) UnmarshalBinary(data []byte) error {
	rd := bytes.NewReader(data)

	_, err := v.Signature.ReadFrom(rd)
	if err != nil {
		return fmt.Errorf("failed to read signature: %w", err)
	}

	var status int32
	if err := binary.Read(rd, binary.LittleEndian, &status); err != nil {
		return fmt.Errorf("failed to read ack status: %w", err)
	}
	v.AckStatus = AckStatus(status)

	if v.AckStatus == AckStatusDiverge {
		var appHash Hash
		if _, err := io.ReadFull(rd, appHash[:]); err != nil {
			return fmt.Errorf("failed to read app hash: %w", err)
		}
		v.AppHash = &appHash
	}

	return nil
}

func (v *VoteInfo) Verify(blkID Hash, appHash Hash) error {
	pubKey, err := crypto.UnmarshalPublicKey(v.Signature.PubKey, v.Signature.PubKeyType)
	if err != nil {
		return fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	var buf bytes.Buffer
	buf.Write(blkID[:])

	switch v.AckStatus {
	case AckStatusDiverge:
		if v.AppHash == nil {
			return errors.New("missing app hash for diverged vote")
		}
		binary.Write(&buf, binary.LittleEndian, true)
		buf.Write((*v.AppHash)[:])
	case AckStatusAgree:
		binary.Write(&buf, binary.LittleEndian, true)
		buf.Write(appHash[:])
	case AckStatusDisagree:
		binary.Write(&buf, binary.LittleEndian, false)
	}

	valid, err := pubKey.Verify(buf.Bytes(), v.Signature.Data)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	if !valid {
		return errors.New("invalid voteInfo signature")
	}

	return nil
}

func SignVote(blkID Hash, ack bool, appHash *Hash, privKey crypto.PrivateKey) (*Signature, error) {
	if privKey == nil {
		return nil, errors.New("nil private key")
	}

	var buf bytes.Buffer
	buf.Write(blkID[:])
	binary.Write(&buf, binary.LittleEndian, ack)
	if ack {
		if appHash == nil {
			return nil, errors.New("missing app hash for ack vote")
		}
		buf.Write(appHash[:])
	}

	sig, err := privKey.Sign(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to sign vote: %w", err)
	}

	return &Signature{
		PubKeyType: privKey.Type(),
		PubKey:     privKey.Public().Bytes(),
		Data:       sig,
	}, nil
}

func (ci *CommitInfo) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	if _, err := buf.Write(ci.AppHash[:]); err != nil {
		return nil, fmt.Errorf("failed to write app hash: %w", err)
	}

	if err := binary.Write(&buf, binary.LittleEndian, int32(len(ci.Votes))); err != nil {
		return nil, fmt.Errorf("failed to write vote count: %w", err)
	}

	for _, v := range ci.Votes {
		voteBytes, err := v.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal vote: %w", err)
		}

		if err := types.WriteBytes(&buf, voteBytes); err != nil {
			return nil, fmt.Errorf("failed to write vote: %w", err)
		}
	}

	puBts, err := ci.ParamUpdates.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal param updates: %w", err)
	}
	if err := types.WriteBytes(&buf, puBts); err != nil {
		return nil, fmt.Errorf("failed to write param updates: %w", err)
	}

	return buf.Bytes(), nil
}

func (ci *CommitInfo) UnmarshalBinary(data []byte) error {
	rd := bytes.NewReader(data)

	if _, err := io.ReadFull(rd, ci.AppHash[:]); err != nil {
		return fmt.Errorf("failed to read app hash: %w", err)
	}

	var voteCount int32
	if err := binary.Read(rd, binary.LittleEndian, &voteCount); err != nil {
		return fmt.Errorf("failed to read vote count: %w", err)
	}

	ci.Votes = make([]*VoteInfo, voteCount)
	for i := range ci.Votes {
		voteBytes, err := types.ReadBytes(rd)
		if err != nil {
			return fmt.Errorf("failed to read vote: %w", err)
		}

		var vote VoteInfo
		if err := vote.UnmarshalBinary(voteBytes); err != nil {
			return fmt.Errorf("failed to unmarshal vote: %w", err)
		}

		ci.Votes[i] = &vote
	}

	puBts, err := types.ReadBytes(rd)
	if err != nil {
		return fmt.Errorf("failed to read param updates: %w", err)
	}
	if err := ci.ParamUpdates.UnmarshalBinary(puBts); err != nil {
		return fmt.Errorf("failed to unmarshal param updates: %w", err)
	}

	return nil
}
