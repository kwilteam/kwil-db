package node

import (
	"fmt"
	"math"

	"github.com/cometbft/cometbft/rpc/core"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

type JoinRequestPool struct {
	joinRequests map[string]*JoinRequest
}

type JoinRequest struct {
	Joiner        string
	Power         int64
	Votes         map[string]bool
	ValidatorSet  map[string]bool
	RequiredVotes int64
	Finalized     bool
}

func NewJoinRequestPool() *JoinRequestPool {
	return &JoinRequestPool{
		joinRequests: make(map[string]*JoinRequest),
	}
}

func ValidatorSetCount() *ctypes.ResultValidators {
	validators, _ := core.Validators(&types.Context{}, nil, nil, nil)
	fmt.Println("Received Validators info: ", validators, " Count: ", validators.Count)
	return validators
}

func (jPool *JoinRequestPool) GetJoinRequest(joiner string, power int64) *JoinRequest {
	if jPool.joinRequests[joiner] != nil {
		return jPool.joinRequests[joiner]
	}

	valSet := ValidatorSetCount()
	cnt := (0.66) * float64(valSet.Count)
	jPool.joinRequests[joiner] = &JoinRequest{
		Joiner:        joiner,
		Votes:         make(map[string]bool),
		ValidatorSet:  make(map[string]bool),
		RequiredVotes: int64(math.Ceil(cnt)),
		Power:         power,
	}

	for _, val := range valSet.Validators {
		jPool.joinRequests[joiner].ValidatorSet[val.Address.String()] = true
	}

	return jPool.joinRequests[joiner]
}

func (jReq *JoinRequest) IsValidator(addr string) bool {
	return jReq.ValidatorSet[addr]
}

func (jPool *JoinRequestPool) AddVote(joiner string, voter string) error {
	// Add vote if the voter is in the validator set
	if jPool.joinRequests[joiner] == nil {
		return fmt.Errorf("joiner %s not found", joiner)
	}

	if jPool.joinRequests[joiner].IsValidator(voter) {
		jPool.joinRequests[joiner].Votes[voter] = true
		jPool.joinRequests[joiner].RequiredVotes--
	}
	jPool.joinRequests[joiner].PrintJoinRequestStats()
	return nil
}

func (jReq *JoinRequest) PrintJoinRequestStats() {
	fmt.Println("Joiner: ", jReq.Joiner, " Power: ", jReq.Power, " RequiredVotes: ", jReq.RequiredVotes, " Finalized: ", jReq.Finalized)
	fmt.Println("Votes: ", jReq.Votes)
	fmt.Println("ValidatorSet: ", jReq.ValidatorSet)
}

func (jPool *JoinRequestPool) GetJoinerPower(joiner string) (int64, error) {
	if jPool.joinRequests[joiner] != nil {
		return jPool.joinRequests[joiner].Power, nil
	} else {
		return 0, fmt.Errorf("joiner %s not found", joiner)
	}
}

func (jPool *JoinRequestPool) AddToValUpdates(joiner string) bool {
	if jPool.joinRequests[joiner] == nil {
		return false
	}
	return jPool.joinRequests[joiner].RequiredVotes == 0 && !jPool.joinRequests[joiner].Finalized
}

func (jPool *JoinRequestPool) RemoveJoinRequest(joiner string) {
	delete(jPool.joinRequests, joiner)
}
