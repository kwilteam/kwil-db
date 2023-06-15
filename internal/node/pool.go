package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmtNode "github.com/cometbft/cometbft/node"
	localClient "github.com/cometbft/cometbft/rpc/client/local"
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
	RequestCh           chan JoinRequest    // Sends requests to the channel, to be consumed by the reactor
	ApprovedVals        *ApprovedValidators // Validators approved by the node
	ApprovedNetworkVals *ApprovedValidators // Validators approved by the majority of nodes in the network
	Join_requests       map[string]*JoinRequestStatus
	BcNode              *cmtNode.Node
}

type JoinRequest struct {
	PubKey crypto.PubKey
	Tx     []byte
}

type JoinRequestStatus struct {
	PubKey crypto.PubKey
	Status int64
	// Should include validator set
	Required_votes int64
	Approved_votes int64
	Rejected_votes int64
	Num_validators int64

	// Validator info on the votes
	Rejected_validators []string // List of cryptographic hashes denoting the nodeIDs of the validators that rejected the request
	Approved_validators []string // List of cryptographic hashes denoting the nodeIDs of the validators that approved the request

	TxHash string // If approved, the hash of the transaction that added the validator

	mu sync.Mutex
}

func NewJoinRequestPool(approvedVals *ApprovedValidators, nwApprovedVals *ApprovedValidators, reqCh chan JoinRequest) *JoinRequestPool {
	return &JoinRequestPool{
		RequestCh:           reqCh,
		ApprovedVals:        approvedVals,
		ApprovedNetworkVals: nwApprovedVals,
		Join_requests:       make(map[string]*JoinRequestStatus),
	}
}

func (pool *JoinRequestPool) AddRequest(request JoinRequest) {
	address := request.PubKey.Address().String()
	fmt.Println("Received join request from: ", address)
	fmt.Println("Received join request from: ", address)
	validators, _ := core.Validators(&types.Context{}, nil, nil, nil)
	fmt.Println("Received Validators info: ", validators, " Count: ", validators.Count)
	numValidators := validators.Count
	delete(pool.Join_requests, address)
	is_validator := pool.ApprovedVals.IsValidator(address)
	pool.Join_requests[address] = &JoinRequestStatus{
		PubKey:              request.PubKey,
		Status:              int64(Initiated),
		Required_votes:      int64(numValidators * 2 / 3),
		Num_validators:      int64(numValidators),
		Approved_votes:      0,
		Rejected_votes:      0,
		Rejected_validators: make([]string, 0),
		Approved_validators: make([]string, 0),
		TxHash:              string(tmhash.Sum(request.Tx)),
		mu:                  sync.Mutex{},
	}
	if is_validator {
		pool.Join_requests[address].Approved_votes += 1
		pool.Join_requests[address].Required_votes -= 1
		pool.Join_requests[address].Approved_validators = append(pool.Join_requests[address].Approved_validators, address)
	} else {
		pool.Join_requests[address].Rejected_votes += 1
		pool.Join_requests[address].Rejected_validators = append(pool.Join_requests[address].Rejected_validators, address)
	}
	fmt.Println(pool.Join_requests[address].Approved_votes, pool.Join_requests[address].Required_votes, pool.Join_requests[address].Approved_validators, pool.Join_requests[address].Rejected_validators)
	pool.RequestCh <- request
	fmt.Println("Added request to requestCh")
}

// Update the status of the request
// Not sure what peer info is available to us
func (pool *JoinRequestPool) AddVote(address string, peerID string, vote bool) Status {
	fmt.Println("Add vote", pool.Join_requests)
	status := pool.Join_requests[address]
	fmt.Println(status, "address", address)
	ready_status := Pending
	status.mu.Lock()
	if vote {
		status.Approved_votes += 1
		status.Required_votes -= 1
		status.Approved_validators = append(status.Approved_validators, peerID)
	} else {
		status.Rejected_votes += 1
		status.Rejected_validators = append(status.Rejected_validators, peerID)
	}
	ready := status.Required_votes < 0
	if ready {
		status.Status = int64(Approved)
		ready_status = Approved
	} else if status.Rejected_votes > status.Num_validators-status.Required_votes {
		status.Status = int64(Rejected)
		ready_status = Rejected
	}
	fmt.Println("Votes: ", status.Approved_votes, status.Required_votes, status.Approved_validators, status.Rejected_validators)
	status.mu.Unlock()

	return ready_status
}

func (pool *JoinRequestPool) AddHash(address string, hash string) {
	pool.Join_requests[address].mu.Lock()
	defer pool.Join_requests[address].mu.Unlock()
	pool.Join_requests[address].TxHash = hash
}

func (pool *JoinRequestPool) SetStatus(address string, status Status) {
	pool.Join_requests[address].mu.Lock()
	defer pool.Join_requests[address].mu.Unlock()
	pool.Join_requests[address].Status = int64(status)

}

func (pool *JoinRequestPool) GetStatus(address string) *JoinRequestStatus {
	pool.Join_requests[address].mu.Lock()
	defer pool.Join_requests[address].mu.Unlock()
	return pool.Join_requests[address]
}

func (pool *JoinRequestPool) IsNodeValidator() bool {
	bcClient := localClient.New(pool.BcNode)
	validators, _ := bcClient.Validators(context.Background(), nil, nil, nil)
	fmt.Printf("Current Validators: %+v\n", validators)
	nodeAddr, _ := pool.BcNode.PrivValidator().GetPubKey()
	fmt.Printf("Node address: %s    %v\n", nodeAddr.Address().String(), nodeAddr)
	for _, val := range validators.Validators {
		addr := fmt.Sprintf("%s", val.Address)
		fmt.Println(addr, nodeAddr.Address().String())
		if addr == nodeAddr.Address().String() {
			fmt.Println("Address matches")
			return true
		}
	}
	return false
}
