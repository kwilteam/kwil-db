package tracking

// TODO: determine if historical state changes and times changed are needed
type Info interface {

	// ID returns the request ID
	ID() string

	// Idempotent_Key returns the source idempotent key
	// to use for finding this message during processing,
	// and for de-duplicating messages in the target Db
	// Note: this addresses de-duplication of messages
	// submitted the actual source producer (e.g.,
	// the client).
	Idempotent_Key() string

	// RequestTime returns the request time
	RequestTime() int64

	// UpdatedTime returns the last updated time
	UpdatedTime() int64

	// GetSource string returns the request user
	GetSource() string

	// Status returns the request status
	Status() Status
}
