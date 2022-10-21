package composer

import (
	"kwil/x/async"
	"kwil/x/svcx/tracking"
)

type Response async.Task[tracking.ID]

// Message >> Need a better handle on the message types
type Message struct {
	Type          uint64
	SourceId      string
	CorrelationId string
	Payload       any
}
