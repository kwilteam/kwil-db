package service

import (
	"context"

	cm "kwil/x/common"
	cfg "kwil/x/common/host/config"
)

type ServiceFactory func() Service

type ServiceContext interface {
	context.Context

	GetServiceById(id int32) (Service, error)
	GetServiceByName(name string) (Service, error)
}

type ServiceIdentity struct {
	id   int32
	name string
}

type Configurable interface {
	Configure(config cfg.Config) error
}

type Initializeable interface {
	Initialize(ctx ServiceContext) error
}

type Service interface {
	Identity() ServiceIdentity
}

type ClosableService interface {
	Service
	cm.Closeable
}

type BackgroundService interface {
	Service

	// Start the service.
	Start(ctx ServiceContext) error

	// Shutdown the service.
	Shutdown()

	// Returns the service's status.
	IsRunning() bool

	// Waits for service shutdown.
	AwaitShutdown()
}
