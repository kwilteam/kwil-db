package composer

import (
	"context"
	"kwil/archive/svcx/tracking"
	"kwil/x/async"
)

type emit_track_runner struct {
	pub     *emit_track
	ctx     context.Context
	message *Message
}

func newEmitTrackRunner(p *emit_track, ctx context.Context, message *Message) *emit_track_runner {
	return &emit_track_runner{p, ctx, message}
}

func (p *emit_track_runner) emit_and_track() Response {
	task := async.ComposeA(p._emit, p._track)
	return async.Map(task, p._map)
}

func (p *emit_track_runner) _emit() async.Action {
	return p.pub.
		emitter.SendT(p.ctx, p.pub.getTopic(p.message), p.message)
}

func (p *emit_track_runner) _track() async.Task[tracking.Item] {
	return p.pub.
		tracker.Submit(p.ctx, p.message.SourceId, p.message.CorrelationId)
}

func (p *emit_track_runner) _map(item tracking.Item) tracking.ID {
	return item.ID()
}
