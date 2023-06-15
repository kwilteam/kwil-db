package node

import (
	"fmt"
	"sync"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmtNode "github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/rpc/core"
	"github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

type Status int

const (
	Initiated Status = 0
	Pending   Status = 1
	Approved  Status = 2
	Rejected  Status = 3
)

type JoinRequestPool struct {
	RequestCh     chan JoinRequest // Sends requests to the channel, to be consumed by the reactor
	ApprovedVals  *ApprovedValidators
	join_requests map[string]*JoinRequestStatus
	BcNode        *cmtNode.Node
}

type JoinRequest struct {
	PubKey crypto.PubKey
	Tx     []byte
}

type JoinRequestStatus struct {
	pubKey crypto.PubKey
	status int64
	// Should include validator set
	required_votes int64
	approved_votes int64
	rejected_votes int64
	num_validators int64

	// Validator info on the votes
	rejected_validators []string // List of cryptographic hashes denoting the nodeIDs of the validators that rejected the request
	approved_validators []string // List of cryptographic hashes denoting the nodeIDs of the validators that approved the request

	TxHash string // If approved, the hash of the transaction that added the validator

	mu sync.Mutex
}

func NewJoinRequestPool(approvedVals *ApprovedValidators, reqCh chan JoinRequest) *JoinRequestPool {
	return &JoinRequestPool{
		RequestCh:     reqCh,
		ApprovedVals:  approvedVals,
		join_requests: make(map[string]*JoinRequestStatus),
	}
}

func (pool *JoinRequestPool) AddRequest(request JoinRequest) {
	address := request.PubKey.Address().String()
	fmt.Println("Received join request from: ", address)
	validators, _ := core.Validators(&types.Context{}, nil, nil, nil)
	fmt.Println("Received Validators info: ", validators)

	/* var found bool = false
	for _, val := range validators.Validators {
		if string(val.Address) == address {
			found = true
			break
		}
	}

	if !found {
		fmt.Println("Node is not a validator - Not broadcasting join request")
		return
	} */

	numValidators := validators.Count
	delete(pool.join_requests, address)
	is_validator := pool.ApprovedVals.IsValidator(address)
	pool.join_requests[address] = &JoinRequestStatus{
		pubKey:              request.PubKey,
		status:              int64(Initiated),
		required_votes:      int64(numValidators * 2 / 3),
		num_validators:      int64(numValidators),
		approved_votes:      0,
		rejected_votes:      0,
		rejected_validators: make([]string, 0),
		approved_validators: make([]string, 0),
		TxHash:              string(tmhash.Sum(request.Tx)),
		mu:                  sync.Mutex{},
	}
	if is_validator {
		pool.join_requests[address].approved_votes += 1
		pool.join_requests[address].required_votes -= 1
		pool.join_requests[address].approved_validators = append(pool.join_requests[address].approved_validators, address)
	}
	fmt.Println(pool.join_requests[address].approved_votes, pool.join_requests[address].required_votes, pool.join_requests[address].approved_validators, pool.join_requests[address].rejected_validators)
	pool.RequestCh <- request
	fmt.Println("Added request to requestCh")
}

// Update the status of the request
// Not sure what peer info is available to us
func (pool *JoinRequestPool) AddVote(address string, peerID string, vote bool) Status {
	fmt.Println("Add vote", pool.join_requests)
	status := pool.join_requests[address]
	fmt.Println(status, "address", address)
	ready_status := Pending
	status.mu.Lock()
	if vote {
		status.approved_votes += 1
		status.required_votes -= 1
		status.approved_validators = append(status.approved_validators, peerID)
	} else {
		status.rejected_votes += 1
		status.rejected_validators = append(status.rejected_validators, peerID)
	}
	ready := status.required_votes < 0
	if ready {
		status.status = int64(Approved)
		ready_status = Approved
	} else if status.rejected_votes > status.num_validators-status.required_votes {
		status.status = int64(Rejected)
		ready_status = Rejected
	}
	fmt.Println("Votes: ", status.approved_votes, status.required_votes, status.approved_validators, status.rejected_validators)
	status.mu.Unlock()

	return ready_status
}

func (pool *JoinRequestPool) AddHash(address string, hash string) {
	pool.join_requests[address].mu.Lock()
	pool.join_requests[address].TxHash = hash
	pool.join_requests[address].mu.Unlock()
}

func (pool *JoinRequestPool) SetStatus(address string, status Status) {
	pool.join_requests[address].mu.Lock()
	pool.join_requests[address].status = int64(status)
	pool.join_requests[address].mu.Unlock()
}
