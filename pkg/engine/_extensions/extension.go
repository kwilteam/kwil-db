package extensions

import (
	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/engine/sqldb"
)

type Extension struct {
	Name         string
	Tables       []*dto.Table
	Initialize   InitializeFunc
	RunCron      RunCronFunc
	GetHeader    GetHeaderFunc
	DecideHeader DecideHeaderFunc
}

// Header is arbitrary data to come to consensus on, which will then be used to run the cron job.
type Header []byte

func (h *Header) Bytes() []byte {
	return *h
}

// InitializeFunc contains arbitrary initialization logic to be run on startup.
// This should serve as an extension "constructor".
type InitializeFunc func(Datastore) error

// RunCronFunc runs a cron job to update the extension.
type RunCronFunc func(Header) error

// GetHeaderFunc returns the header to come to consensus on.
type GetHeaderFunc func() (Header, error)

// DecideHeaderFunc takes a set of headers and decides on a single header to use.
// It should be deterministic.
type DecideHeaderFunc func(...Header) (Header, error)

type Datastore sqldb.Datastore
