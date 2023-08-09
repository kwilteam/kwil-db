package node

import (
	"encoding/json"
	"fmt"

	cmtCrypto "github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/kwilteam/kwil-db/pkg/serialize"
)

type ValidatorsInfo struct {
	ValidatorApprovalVotes map[string][]string // All the approval votes received till now
	//ValAddrToPubKeyMap     map[string]pc.PublicKey // Map of validator address to public key of current validators
	FinalizedValidators map[string]bool // Map of validators who have been approved with majority of votes
}

func NewValidatorsInfo() *ValidatorsInfo {
	return &ValidatorsInfo{
		ValidatorApprovalVotes: make(map[string][]string),
		//ValAddrToPubKeyMap:     make(map[string]pc.PublicKey),
		FinalizedValidators: make(map[string]bool),
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

func UnmarshalValidator(payload []byte) (*serialize.Validator, error) {
	var validator serialize.Validator
	err := json.Unmarshal(payload, &validator)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal validator: %w", err)
	}

	return &validator, nil
}

func UnmarshalPublicKey(addr []byte) (cmtCrypto.PubKey, error) {
	var publicKey cmtCrypto.PubKey
	key := fmt.Sprintf(`{"type":"tendermint/PubKeyEd25519","value":%s}`, string(addr))
	fmt.Println("Key:", key)

	err := cmtjson.Unmarshal([]byte(key), &publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal validator public key: %w", err)
	}
	fmt.Println("publicKey: ", publicKey)
	return publicKey, nil
}
