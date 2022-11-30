package processor

import (
	"fmt"
	"kwil/_archive/svcx/messaging/mx"
	"kwil/_archive/svcx/wallet"
	dt "kwil/x/deposits/types"
	"log"
)

func AsMessageTransform(pr Processor) wallet.MessageTransform {
	return wallet.SyncTransform(func(msg *mx.RawMessage) (*mx.RawMessage, error) {
		// determine message type
		mt := msg.Value[1]
		switch mt {
		default:
			return nil, fmt.Errorf("unknown message type: %v", msg.Value)
		case 0x0:
			// deposit
			deposit, err := dt.Deserialize[*dt.Deposit](msg.Value)
			if err != nil {
				return nil, err
			}

			err = pr.ProcessDeposit(deposit)
			if err != nil {
				return nil, err
			}
			log.Println("deposit processed")
		case 0x01:
			// withdrawal request
			wdr, err := dt.Deserialize[*dt.WithdrawalRequest](msg.Value)
			if err != nil {
				return nil, err
			}

			err = pr.ProcessWithdrawalRequest(wdr)
			if err != nil {
				return nil, err
			}
		case 0x02:
			// withdrawal confirmation
			wdc, err := dt.Deserialize[*dt.WithdrawalConfirmation](msg.Value)
			if err != nil {
				return nil, err
			}

			pr.ProcessWithdrawalConfirmation(wdc)
		case 0x03:
			// End Of Block
			eob, err := dt.Deserialize[*dt.EndBlock](msg.Value)
			if err != nil {
				return nil, err
			}

			// TODO: confirm the error does not need to be handled
			pr.ProcessEndBlock(eob)
		case 0x04:
			// Spend
			spend, err := dt.Deserialize[*dt.Spend](msg.Value)
			if err != nil {
				return nil, err
			}

			err = pr.ProcessSpend(spend)
			if err != nil {
				return nil, err
			}
		}

		return msg, nil
	})
}
