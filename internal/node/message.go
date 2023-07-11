package node

import (
	"fmt"

	"github.com/cometbft/cometbft/p2p"
	"github.com/cosmos/gogoproto/proto"
)

var _ p2p.Wrapper = &ValidatorJoinRequestVote{}
var _ p2p.Wrapper = &ValidatorJoinResponseVote{}

func (m *ValidatorJoinRequestVote) Wrap() proto.Message {
	sm := &Message{}
	sm.Sum = &Message_ValidatorJoinRequestVote{ValidatorJoinRequestVote: m}
	return sm
}

func (m *ValidatorJoinResponseVote) Wrap() proto.Message {
	sm := &Message{}
	sm.Sum = &Message_ValidatorJoinResponseVote{ValidatorJoinResponseVote: m}
	return sm
}

func (m *Message) Unwrap() (proto.Message, error) {
	switch msg := m.Sum.(type) {
	case *Message_ValidatorJoinRequestVote:
		return m.GetValidatorJoinRequestVote(), nil

	case *Message_ValidatorJoinResponseVote:
		return m.GetValidatorJoinResponseVote(), nil

	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}
