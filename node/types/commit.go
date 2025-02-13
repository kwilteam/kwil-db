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

var (
	SerializationByteOrder = types.SerializationByteOrder
)

type NodeStatus struct {
	Role            string                   `json:"role"`
	CatchingUp      bool                     `json:"catching_up"`
	CommittedHeader *types.BlockHeader       `json:"committed_header"`
	CommitInfo      *CommitInfo              `json:"commit_info"`
	Params          *types.NetworkParameters `json:"params"`
}

// CommitInfo includes the information about the commit of the block.
// Such as the signatures of the validators aggreeing to the block.
type CommitInfo struct {
	AppHash          Hash
	Votes            []*VoteInfo
	ParamUpdates     types.ParamUpdates
	ValidatorUpdates []*types.Validator
}

type AckStatus int

// This is how leader interprets the vote(AckRes) into the VoteInfo.
// Nack                    -- Rejected
// Ack + same AppHash      -- Agreed
// Ack + different AppHash -- Forked

const (
	// Rejected means the validator did not accept the proposed block and
	// responded with a NACK. This can occur due to issues like apphash mismatch,
	// validator set mismatch, consensus params mismatch, merkle root mismatch, etc.
	Rejected AckStatus = iota
	// Agreed means the validator accepted the proposed block and
	// computed the same AppHash as the leader after processing the block.
	Agreed
	// Forked means the validator accepted the proposed block and
	// successfully processed it, but diverged after processing the block.
	// The leader identifies this from the app hash mismatch in the vote.
	Forked
)

func (ack *AckStatus) String() string {
	switch *ack {
	case Rejected:
		return "rejected"
	case Agreed:
		return "agreed"
	case Forked:
		return "forked"
	default:
		return "unknown"
	}
}

func (ack AckStatus) WasAck() bool {
	switch ack {
	case Agreed, Forked:
		return true
	default: // Rejected
		return false
	}
}

// VoteInfo represents the leader's interpretation of the AckRes vote received from the validator.
// This only includes the votes that influenced the commit decision of the block. It does not include
// the feedback votes for an already committed block such as OutOfSync Vote etc.
// Validators and sentry nodes use this information from the CommitInfo to verify that the
// committed block state was agreed upon by the majority of the validators from the validator set.
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

func (sig *Signature) Bytes() []byte {
	var buf bytes.Buffer
	sig.WriteTo(&buf)
	return buf.Bytes()
}

func DecodeSignature(data []byte) (*Signature, error) {
	sig := &Signature{}
	if _, err := sig.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *Signature) WriteTo(w io.Writer) (int64, error) {
	cw := utils.NewCountingWriter(w)
	// PubKeyType
	if _, err := s.PubKeyType.WriteTo(cw); err != nil {
		return cw.Written(), err
	}

	// PubKey Length
	if err := types.WriteCompactBytes(cw, s.PubKey); err != nil {
		return cw.Written(), err
	}

	// Signature Data Length
	if err := types.WriteCompactBytes(cw, s.Data); err != nil {
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

	pubKey, err := types.ReadCompactBytes(cr)
	if err != nil {
		return cr.ReadCount(), fmt.Errorf("failed to read public key: %w", err)
	}
	s.PubKey = pubKey

	sig, err := types.ReadCompactBytes(cr)
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

	if v.AckStatus == Forked {
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

	if v.AckStatus == Forked {
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
	case Forked:
		if v.AppHash == nil {
			return errors.New("missing app hash for diverged vote")
		}
		binary.Write(&buf, binary.LittleEndian, true)
		buf.Write((*v.AppHash)[:])
	case Agreed:
		binary.Write(&buf, binary.LittleEndian, true)
		buf.Write(appHash[:])
	case Rejected:
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

	if _, err := buf.Write(binary.AppendUvarint(nil, uint64(len(ci.Votes)))); err != nil {
		return nil, fmt.Errorf("failed to write vote count: %w", err)
	}

	for _, v := range ci.Votes {
		voteBytes, err := v.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal vote: %w", err)
		}

		if err := types.WriteCompactBytes(&buf, voteBytes); err != nil {
			return nil, fmt.Errorf("failed to write vote: %w", err)
		}
	}

	// Param Updates
	puBts, err := ci.ParamUpdates.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal param updates: %w", err)
	}
	if err := types.WriteCompactBytes(&buf, puBts); err != nil {
		return nil, fmt.Errorf("failed to write param updates: %w", err)
	}

	// Validator Updates
	if _, err := buf.Write(binary.AppendUvarint(nil, uint64(len(ci.ValidatorUpdates)))); err != nil {
		return nil, fmt.Errorf("failed to write validator update count: %w", err)
	}
	for _, val := range ci.ValidatorUpdates {
		valBts, err := val.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal validator: %w", err)
		}
		if err := types.WriteCompactBytes(&buf, valBts); err != nil {
			return nil, fmt.Errorf("failed to write validator: %w", err)
		}
	}

	return buf.Bytes(), nil
}

func (ci *CommitInfo) UnmarshalBinary(data []byte) error {
	rd := bytes.NewReader(data)

	if _, err := io.ReadFull(rd, ci.AppHash[:]); err != nil {
		return fmt.Errorf("failed to read app hash: %w", err)
	}

	voteCount, err := binary.ReadUvarint(rd)
	if err != nil {
		return fmt.Errorf("failed to read vote count: %w", err)
	}

	ci.Votes = make([]*VoteInfo, voteCount)
	for i := range ci.Votes {
		voteBytes, err := types.ReadCompactBytes(rd)
		if err != nil {
			return fmt.Errorf("failed to read vote: %w", err)
		}

		var vote VoteInfo
		if err := vote.UnmarshalBinary(voteBytes); err != nil {
			return fmt.Errorf("failed to unmarshal vote: %w", err)
		}

		ci.Votes[i] = &vote
	}

	puBts, err := types.ReadCompactBytes(rd)
	if err != nil {
		return fmt.Errorf("failed to read param updates: %w", err)
	}
	if err := ci.ParamUpdates.UnmarshalBinary(puBts); err != nil {
		return fmt.Errorf("failed to unmarshal param updates: %w", err)
	}

	valCount, err := binary.ReadUvarint(rd)
	if err != nil {
		return fmt.Errorf("failed to read validator update count: %w", err)
	}
	ci.ValidatorUpdates = make([]*types.Validator, valCount)
	for i := range ci.ValidatorUpdates {
		valBts, err := types.ReadCompactBytes(rd)
		if err != nil {
			return fmt.Errorf("failed to read validator: %w", err)
		}

		val := &types.Validator{}
		if err := val.UnmarshalBinary(valBts); err != nil {
			return fmt.Errorf("failed to unmarshal validator: %w", err)
		}

		ci.ValidatorUpdates[i] = val
	}

	return nil
}
