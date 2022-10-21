package composer

import (
	"context"
	"kwil/x/svcx/mapper"
	"kwil/x/svcx/messaging/pub"
	"kwil/x/svcx/tracking"
)

type emit_track struct {
	emitter  pub.Emitter[*Message]
	tracker  tracking.Service
	resolver mapper.TopicMapper
}

func (p *emit_track) Submit(ctx context.Context, message *Message) Response {
	return newEmitTrackRunner(p, ctx, message).emit_and_track()
}

func (p *emit_track) getTopic(message *Message) string {
	return p.resolver.GetTopic(p.asMsgCtx(message))
}

func (p *emit_track) asMsgCtx(message *Message) mapper.MessageContext {
	panic("implement me")
}
