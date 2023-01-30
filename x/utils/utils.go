package utils

// joinBytes is a helper function to join multiple byte slices into one
func JoinBytes(s ...[]byte) []byte {
	n := 0
	for _, v := range s {
		n += len(v)
	}

	b, i := make([]byte, n), 0
	for _, v := range s {
		i += copy(b[i:], v)
	}
	return b
}

// JoinBytesSize joins multiple byte slices into one with a specific size known beforehand
func JoinBytesSize(size int, s ...[]byte) []byte {
	b, i := make([]byte, size), 0
	for _, v := range s {
		i += copy(b[i:], v)
	}
	return b
}
