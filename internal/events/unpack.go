package events

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

// This function takes a log from ethereum and unpacks it into a deposit struct
func (ef *EventFeed) unpackDeposit(vLog ethTypes.Log) (*DepositEvent, error) {
<<<<<<< HEAD
	abi := ef.conf.GetContractABI()
=======
	abi := ef.Config.ClientChain.GetContractABI()
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	y, _ := abi.Unpack("Deposit", vLog.Data)
	dep := Deposit{
		Caller: y[0].(common.Address),
		Target: y[1].(common.Address),
		Amount: y[2].(*big.Int),
	}
	err := abi.UnpackIntoInterface(&dep, "Deposit", vLog.Data)
	if err != nil {
		return nil, fmt.Errorf("error unpacking deposit event: %s", err)
	}

	return &DepositEvent{
		Name:   "Deposit",
		Height: big.NewInt(int64(vLog.BlockNumber)),
		Data:   &dep,
		Tx:     vLog.TxHash.Bytes(),
	}, nil
}

type Deposit struct {
	Caller common.Address //`abi:"caller"`
	Target common.Address //`abi:"target"`
	Amount *big.Int       //`abi:"amount"`
}

func (ef *EventFeed) parseEvent(vLog ethTypes.Log) (Event, error) {
	topic := vLog.Topics[0]
	event := ef.Topics[topic]
	switch event.Name {
	default:
		return nil, fmt.Errorf("unknown event type: %s", event.Name)
	case "Deposit":
		return ef.unpackDeposit(vLog)
	}
}
