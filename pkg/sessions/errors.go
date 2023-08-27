package sessions

import (
	"errors"
)

var (
	ErrSessionInProgress   = errors.New("session already in progress")
	ErrNoSessionInProgress = errors.New("no session in progress, cannot commit")
	ErrCommitPhase         = errors.New("incorrect commit phase")
	ErrMissingBegin        = errors.New("missing begin record")
	ErrBeginCommit         = errors.New("error beginning atomic commit")
	ErrEndCommit           = errors.New("error ending atomic commit")
	ErrBeginApply          = errors.New("error beginning apply")
	ErrApply               = errors.New("error applying changes")
	ErrEndApply            = errors.New("error ending apply")
	ErrID                  = errors.New("error generating session ID")
	ErrAlreadyRegistered   = errors.New("committable already registered")
	ErrUnknownCommittable  = errors.New("unknown committable")
	ErrClosed              = errors.New("session closed")
)

// wrapError wraps an error with a message.
func wrapError(err error, msg error) error {
	return errors.Join(err, msg)
}
