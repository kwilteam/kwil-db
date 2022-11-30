package tracking

// Item
// TODO: determine if historical state changes and times changed are needed
type Item interface {
	// ID returns the item ID
	ID() ID

	// CorrelationId returns the source idempotent key
	// to use for finding this message during processing,
	// and for de-duplicating messages in the target Db
	// Note: this addresses de-duplication of messages
	// submitted by the actual source producer (e.g.,
	// the client).
	CorrelationId() string

	// GetSourceIdentity string returns the item source
	// identity (e.g., the client id, etc)
	GetSourceIdentity() string

	// Created returns the request time
	Created() int64

	// Updated returns the last updated time
	Updated() int64

	// Status returns the item status
	Status() Status
}
