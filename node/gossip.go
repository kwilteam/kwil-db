package node

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const (
	TopicACKs     = "acks"
	TopicReset    = "reset"
	TopicDiscReq  = "discovery_request"
	TopicDiscResp = "discovery_response"
)

func subTopic(_ context.Context, ps *pubsub.PubSub, topic string) (*pubsub.Topic, *pubsub.Subscription, error) {
	th, err := ps.Join(topic)
	if err != nil {
		return nil, nil, err
	}

	sub, err := th.Subscribe()
	if err != nil {
		return nil, nil, err
	}
	return th, sub, nil
}

// WARNING: startTxGossip is an OUT OF DATE gossipsub transaction gossip system.
// Distinct mempool and confirmed tx index have since been added.

/* func (n *Node) startTxGossip(ctx context.Context, ps *pubsub.PubSub) error {
	topicTx, subTx, err := subTxs(ctx, ps)
	if err != nil {
		return err
	}

	subCanceled := make(chan struct{})

	n.wg.Add(1)
	go func() {
		defer func() {
			<-subCanceled
			topicTx.Close()
			n.wg.Done()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}

			txHash := randBytes(32)
			txid := hex.EncodeToString(txHash)
			n.mp.Store(types.Hash(txHash), randBytes(10))
			fmt.Printf("announcing txid %x\n", txid)
			err := topicTx.Publish(ctx, []byte(txid))
			if err != nil {
				fmt.Println("Publish:", err)
				return
			}
		}
	}()

	me := n.host.ID()

	go func() {
		defer close(subCanceled)
		defer subTx.Cancel()
		for {
			txMsg, err := subTx.Next(ctx)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					n.log.Errorf("Stopping TX gossip!", "error", err)
				}
				return
			}

			if string(txMsg.From) == string(me) {
				// n.log.Infoln("message from me ignored")
				continue
			}

			txHash := types.Hash(txMsg.Data)
			txid := txHash.String()
			fromPeerID := txMsg.GetFrom()

			have := n.mp.Get(txHash) != nil // danger conversion
			n.log.Infof("received tx msg from %v (rcvd from %s), data = %x, already have = %v\n",
				txMsg.GetFrom(), txMsg.ReceivedFrom, txMsg.Message.Data, have)
			if have {
				continue
			}

			// Now we use getTx with the ProtocolIDTransaction stream
			n.log.Info("Fetching tx", "hash", txid)
			if n.host.Network().Connectedness(fromPeerID) != network.Connected {
				n.log.Infof("DELAY for fetch, gossip from non-peer")
				time.Sleep(200 * time.Millisecond)
			}
			txRaw, err := n.getTxWithRetry(ctx, txHash, 500*time.Millisecond, 10)
			if err != nil {
				n.log.Errorf("unable to retrieve tx %v: %v", txid, err)
				continue
			}

			n.mp.Store(types.Hash(txMsg.Data), txRaw) // danger conversion

			// txMsg.ID
			// txMsg.ReceivedFrom
			// txMsg.ValidatorData
			// txMsg.Message.Signature
		}
	}()

	return nil
}
*/
