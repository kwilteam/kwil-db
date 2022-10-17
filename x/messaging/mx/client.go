package mx

import "kwil/x"

type ClientType int

const (
	Emitter_Type ClientType = iota
	Receiver_Type
)

type Client interface {
	GetClientType() ClientType
	IsClosed() bool
	Close() bool
	OnClosed() <-chan x.Void
}
