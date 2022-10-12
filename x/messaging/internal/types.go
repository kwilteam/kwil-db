package internal

import "kwil/x"

type Closable interface {
	ID() int
	Close()
	OnClosed() <-chan x.Void
}
