package composer

import (
	"kwil/archive/svcx/tracking"
	"kwil/x/async"
)

type Response async.Task[tracking.ID]

// Message >> Need a better handle on the message types
type Message struct {
	Type          uint64
	SourceId      string
	CorrelationId string
	Payload       any
}
