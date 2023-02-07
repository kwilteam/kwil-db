package chainsync

type chunkRange [2]int64

func splitBlocks(start, end, chunkSize int64) []chunkRange {
	if start == end {
		return []chunkRange{{start, start}}
	}
	var chunks []chunkRange
	for i := start; i < end; i += chunkSize {
		chunkEnd := i + chunkSize
		if chunkEnd > end {
			chunkEnd = end
		}
		chunks = append(chunks, chunkRange{i, chunkEnd - 1})
	}

	if chunks[len(chunks)-1][1] != end {
		chunks[len(chunks)-1][1] = end
	}
	return chunks
}
