package node

import (
	pc "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

type ValidatorsInfo struct {
	ValidatorApprovalVotes map[string][]string     // All the approval votes received till now
	ValAddrToPubKeyMap     map[string]pc.PublicKey // Map of validator address to public key of current validators
	FinalizedValidators    map[string]bool         // Map of validators who have been approved with majority of votes
}

func NewValidatorsInfo() *ValidatorsInfo {
	return &ValidatorsInfo{
		ValidatorApprovalVotes: make(map[string][]string),
		ValAddrToPubKeyMap:     make(map[string]pc.PublicKey),
		FinalizedValidators:    make(map[string]bool),
	}
}

func (vInfo *ValidatorsInfo) AddApprovedValidator(joiner string, approver string) {
	vInfo.ValidatorApprovalVotes[joiner] = append(vInfo.ValidatorApprovalVotes[joiner], approver)
}

func (vInfo *ValidatorsInfo) IsJoinerApproved(joiner string, approver string) bool {
	for _, val := range vInfo.ValidatorApprovalVotes[joiner] {
		if val == approver {
			return true
		}
	}
	return false
}
