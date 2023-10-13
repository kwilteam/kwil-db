package datasets

import "errors"

var (
	ErrInsufficientFee      = errors.New("insufficient fee")
	ErrAuthenticationFailed = errors.New("authentication failed")
)
