package host

import (
	"errors"
)

var (
	// ErrInvalidServiceIdentity is returned when a service is not found.
	ErrInvalidServiceIdentity = errors.New("invalid service identity")

	// ErrServiceNotFound is returned when a service is not found.
	ErrServiceIsNil = errors.New("service factory returned nil")

	// ErrServiceNotFound is returned when a service is not found.
	ErrServiceNotFound = errors.New("service not found")

	// ErrServiceAlreadyRegistered is returned when a service is already registered.
	ErrServiceAlreadyRegistered = errors.New("service already registered")

	// ErrHostAlreadyRunning is returned when the host is already running.
	ErrHostAlreadyRunning = errors.New("host already running")

	// ErrHostNotRunning is returned when the host is not running.
	ErrHostNotRunning = errors.New("host not running")

	// ErrHostShutdown is returned when the host is shutdown.
	ErrHostShutdown = errors.New("host shutdown")

	// ErrHostErrored is returned when the host is in an errored state.
	ErrHostErrored = errors.New("host in errored state")

	// ErrHostRegistrationClosed is returned when the host registration is closed.
	ErrHostRegistrationClosed = errors.New("host registration closed")
)
