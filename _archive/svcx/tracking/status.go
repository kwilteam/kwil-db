package tracking

type Status int32

const (
	// StatusPending is the status when the request is pending
	StatusPending Status = 1
	// StatusProcessing is the status when the request is being processed
	StatusProcessing Status = 2
	// StatusComplete is the status when the request is complete
	StatusComplete Status = 3
	// StatusFailed is the status when the request has failed
	StatusFailed Status = 4
)
