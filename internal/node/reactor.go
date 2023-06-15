package node

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	localClient "github.com/cometbft/cometbft/rpc/client/local"
	// nodepb "github.com/kwilteam/kwil-db/api/protobuf/node/v1"
)

const (
	// NodeJoin Channel is used to make decisions about whether to admit a node as a validator or not
	NodeJoinChannel = byte(0x70)
	maxJoinRequests = 100 // TODO: make this configurable?
)

type Reactor struct {
	p2p.BaseReactor

	// TODO: add fields for state
	pool      *JoinRequestPool
	RequestCh chan JoinRequest // Receives requests from the pool
	txs       map[string][]byte
	// LocalABCIClient to send a Validator Join transaction to the ABCI Application

	Wg sync.WaitGroup
}

func NewReactor(approvedVals *ApprovedValidators) *Reactor {
	requestsCh := make(chan JoinRequest, maxJoinRequests)
	pool := NewJoinRequestPool(approvedVals, requestsCh)
	nodeR := &Reactor{
		pool:      pool,
		RequestCh: requestsCh,
		txs:       make(map[string][]byte),
	}
	nodeR.BaseReactor = *p2p.NewBaseReactor("Node", nodeR)
	return nodeR
}

func (r *Reactor) OnStart() error {
	// go-routine that consumes requests on channel RequestCh
	//r.Wg.Add(1)
	//go r.joinRequestRoutine()

	return nil
}

func (r *Reactor) SetLogger(l log.Logger) {

	r.BaseService.Logger = l
}

func (r *Reactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  NodeJoinChannel,
			Priority:            3,
			SendQueueCapacity:   1000,
			RecvBufferCapacity:  50 * 4096,
			RecvMessageCapacity: 104857605,
			// MessageType:         &nodepb.Message{},
			MessageType: &Message{},
		},
	}
}

// InitPeer implements Reactor by creating a state for the peer.
func (r *Reactor) InitPeer(peer p2p.Peer) p2p.Peer {
	return peer
}

func (r *Reactor) AddPeer(peer p2p.Peer) {
	// TODO: N/A for now
}

func (r *Reactor) RemovePeer(peer p2p.Peer, reason interface{}) {
	// TODO: N/A for now
}

func (r *Reactor) ReceiveEnvelope(e p2p.Envelope) {
	if !r.IsRunning() {
		fmt.Println("Node Reactor is not running, ignoring message")
		return
	}
	// TODO:
	/* TODO: Check that the message is received from the current validator set, else ignore the request and maybe block the peer?*/
	fmt.Println("Node Reactor Receive msg: ", "e.Src", e.Src, "chID", e.ChannelID, "msg", e.Message, "type", reflect.TypeOf(e.Message))
	switch msg := e.Message.(type) {
	case *ValidatorJoinRequestVote:
		/*
			Received request to vote to join a ndoe as a validator
			1. Check if the node is in the approved list of a validator
			2. Based on that, send a vote to the sender - not a broadcast
		*/
		address := string(msg.ValidatorAddress)
		fmt.Println("Received request to vote for Validator ", address, "from", e.Src.ID())
		vote := r.pool.ApprovedVals.IsValidator(address)
		e.Src.SendEnvelope(p2p.Envelope{
			ChannelID: NodeJoinChannel,
			Message: &ValidatorJoinResponseVote{
				ValidatorAddress: msg.ValidatorAddress,
				Accepted:         vote,
			},
		})

	case *ValidatorJoinResponseVote:
		/*
			Received vote for the nodeJoinRequest of a validator
			Count the votes and decide whether to admit the node as a validator or not
			If Validator is in approved state, send a ValidatorJoin transaction to the ABCI Application
		*/
		fmt.Println("Received vote", "vote", msg.Accepted, "from", e.Src.ID())
		if r.pool.AddVote(msg.ValidatorAddress, string(e.Src.ID()), msg.Accepted) == Approved {
			// Send a ValidatorJoin transaction to the ABCI Application
			fmt.Println("Sending ValidatorJoin transaction to ABCI Application")
			tx := r.txs[msg.ValidatorAddress]
			bcClient := localClient.New(r.pool.BcNode)
			res, err := bcClient.BroadcastTxAsync(context.Background(), tx)
			if err != nil {
				fmt.Println("Error broadcasting tx", "err", err, "address", msg.ValidatorAddress)
				return
			}
			fmt.Println("Broadcasted tx to ABCI app", "res", res, "address", msg.ValidatorAddress)
			r.pool.AddHash(msg.ValidatorAddress, string(res.Hash))
		}
	default:
		fmt.Println("Unknown message type", reflect.TypeOf(msg))
	}
}

func (r *Reactor) JoinRequestRoutine() {
	defer r.Wg.Done()
	fmt.Println("Enter joinRequestRoutine")
	for {
		fmt.Println("Waiting for join requests on channel")
		select {
		case request := <-r.RequestCh:
			fmt.Println("Received request to join as a validator", request)
			address := request.PubKey.Address().String()
			r.Switch.BroadcastEnvelope(p2p.Envelope{
				ChannelID: NodeJoinChannel,
				Message: &ValidatorJoinRequestVote{
					ValidatorAddress: address,
				},
			})
			fmt.Println("Broadcasting request to join as a validator", "address", address)
			r.pool.SetStatus(address, Pending)
			r.txs[address] = request.Tx
		case <-r.Quit():
			fmt.Println("Exit Node Reactor")
			return
		}

	}
}

func (r *Reactor) GetPool() *JoinRequestPool {
	return r.pool
}
