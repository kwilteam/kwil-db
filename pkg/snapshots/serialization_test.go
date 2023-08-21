package snapshots_test

import (
	"encoding/json"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/snapshots"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name     string
	snapshot snapshots.Snapshot
}

var (
	testCases = []testCase{
		{
			name: "Snapshot with single chunk",
			snapshot: snapshots.Snapshot{
				Height:     1,
				Format:     1,
				ChunkCount: 1,
				Metadata: snapshots.SnapshotMetadata{
					ChunkHashes: map[uint32][]byte{
						0: []byte("hello"),
					},
					FileInfo: map[string]snapshots.SnapshotFileInfo{
						"file1": {
							Size:     15,
							Hash:     []byte("hash"),
							BeginIdx: 0,
							EndIdx:   0,
						},
					},
				},
			},
		},
		{
			// This tc tests serialization consistency of maps (Go 1.17 onwards iteration order of maps are consistent)
			name: "Snapshot with multiple chunks",
			snapshot: snapshots.Snapshot{
				Height:     1,
				Format:     1,
				ChunkCount: 3,
				Metadata: snapshots.SnapshotMetadata{
					ChunkHashes: map[uint32][]byte{
						10: []byte("hello"),
						21: []byte("hello2"),
						12: []byte("hello3"),
						32: []byte("hello4"),
					},
					FileInfo: map[string]snapshots.SnapshotFileInfo{
						"file1": {
							Size:     15,
							Hash:     []byte("hash"),
							BeginIdx: 0,
							EndIdx:   0,
						},
						"randomfile1": {
							Size:     43,
							Hash:     []byte("hash"),
							BeginIdx: 40,
							EndIdx:   89,
						},
						"file2": {
							Size:     19,
							Hash:     []byte("hash2"),
							BeginIdx: 1,
							EndIdx:   2,
						},
					},
				},
			},
		},
		{
			name: "Snapshot with empty metadata",
			snapshot: snapshots.Snapshot{
				Height:     1,
				Format:     1,
				ChunkCount: 1,
				Metadata: snapshots.SnapshotMetadata{
					ChunkHashes: map[uint32][]byte{},
					FileInfo:    map[string]snapshots.SnapshotFileInfo{},
				},
			},
		},
	}
)

func Test_Snapshot_Serialization(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.MarshalIndent(tc.snapshot, "", "  ")
			assert.NoError(t, err)

			for i := 0; i < 5; i++ {
				var snapshot snapshots.Snapshot
				err = json.Unmarshal(jsonData, &snapshot)
				assert.NoError(t, err)
				assert.EqualValues(t, tc.snapshot, snapshot, "Decoded values doesnt match")
			}
		})
	}
}
