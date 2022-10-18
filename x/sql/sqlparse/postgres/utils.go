package postgres

func makeByte(s string) byte {
	var b byte
	if s == "" {
		return b
	}
	return []byte(s)[0]
}

func makeUint32Slice(in []uint64) []uint32 {
	out := make([]uint32, len(in))
	for i, v := range in {
		out[i] = uint32(v)
	}
	return out
}

func makeString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toPointer(x int) *int {
	return &x
}
