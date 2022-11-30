package composer

import (
	"context"
	"kwil/_archive/svcx/messaging/pub"
	"kwil/_archive/svcx/tracking"
	"kwil/x/cfgx"
)

const SERVICE_ALIAS = "emit-tracking-service" // todo: ensure unique

type EmitTrack interface {
	Submit(ctx context.Context, message *Message) Response
}

func New(cfg cfgx.Config, tracker tracking.Service) (EmitTrack, error) {
	emitter, err := pub.NewEmitterSingleClient[*Message](cfg, &msg_serdes{})
	if err != nil {
		return nil, err
	}

	// load resolver here

	return &emit_track{emitter, tracker, nil}, nil
}
