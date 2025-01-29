package evmsync

import "sync"

// this file contains a thread-safe in-memory cache for the chains that the network cares about.

type listenerInfo struct {
	// mu protects all fields in this struct
	mu sync.RWMutex
}
