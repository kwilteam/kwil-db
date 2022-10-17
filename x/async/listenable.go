package async

// Listenable is an interface that allows for
// combining/composing different Promise-like
// implementations.
type Listenable[T any] interface {
	// IsError will return true if the Result is an error.
	// It will return false if it has not yet completed.
	IsError() bool

	// IsCancelled will return true if the Result is cancelled.
	IsCancelled() bool

	// IsErrorOrCancelled will return true if the Result is
	// an error or cancelled.
	IsErrorOrCancelled() bool

	// IsDone will return true if the Result is complete
	IsDone() bool

	// OnComplete will call the func when the result has
	// been set
	OnComplete(*Continuation[T])
}
