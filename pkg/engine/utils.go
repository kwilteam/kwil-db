package engine

import (
	"sync/atomic"

	"github.com/kwilteam/kwil-db/pkg/engine/utils"
)

// GenerateDBID generates a DBID from a name and owner
func GenerateDBID(name, owner string) string {
	return utils.GenerateDBID(name, owner)
}

// atomicBool is a struct that represents an atomic boolean.
type atomicBool struct {
	val int32 // val is the integer representation of the boolean. 0 for false, 1 for true.
}

// Set updates the value of the boolean atomically.
func (b *atomicBool) Set(value bool) {
	var i int32 = 0
	if value {
		i = 1 // If the input value is true, set i to 1.
	}
	// atomic.StoreInt32 atomically stores i into &b.val.
	atomic.StoreInt32(&b.val, int32(i))
}

// Get retrieves the value of the boolean atomically.
func (b *atomicBool) Get() bool {
	// atomic.LoadInt32 atomically loads &b.val.
	// If the value is not 0, it is interpreted as true.
	return atomic.LoadInt32(&b.val) != 0
}
