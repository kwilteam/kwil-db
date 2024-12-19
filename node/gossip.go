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
