package mx

import "kwil/x"

type ClientType int

const (
	Emitter ClientType = iota
	Receiver
)

type Client interface {
	GetClientType() ClientType
	IsClosed() bool
	Close() bool
	OnClosed() <-chan x.Void
}
