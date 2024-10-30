package node

import (
	"bytes"
	"encoding/binary"
	"io"
	"kwil/node/types"
	"testing"
	"time"
)

func TestBlockProp_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		bp      blockProp
		wantErr bool
	}{
		{
			name: "valid block proposal",
			bp: blockProp{
				Height:    100,
				Hash:      [32]byte{1, 2, 3},
				PrevHash:  [32]byte{4, 5, 6},
				Stamp:     time.Now().Unix(),
				LeaderSig: []byte{7, 8, 9, 10},
			},
			wantErr: false,
		},
		{
			name: "zero values",
			bp: blockProp{
				Height:    0,
				Hash:      [32]byte{},
				PrevHash:  [32]byte{},
				Stamp:     0,
				LeaderSig: []byte{1, 2, 3, 4},
			},
			wantErr: false,
		},
		{
			name: "large signature",
			bp: blockProp{
				Height:    1,
				Hash:      [32]byte{1},
				PrevHash:  [32]byte{2},
				Stamp:     1,
				LeaderSig: bytes.Repeat([]byte{1}, 100),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.bp.MarshalBinary()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var newBp blockProp
			err = newBp.UnmarshalBinary(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if newBp.Height != tt.bp.Height {
					t.Errorf("Height mismatch: got %v, want %v", newBp.Height, tt.bp.Height)
				}
				if newBp.Hash != tt.bp.Hash {
					t.Errorf("Hash mismatch: got %v, want %v", newBp.Hash, tt.bp.Hash)
				}
				if newBp.PrevHash != tt.bp.PrevHash {
					t.Errorf("PrevHash mismatch: got %v, want %v", newBp.PrevHash, tt.bp.PrevHash)
				}
				if newBp.Stamp != tt.bp.Stamp {
					t.Errorf("Stamp mismatch: got %v, want %v", newBp.Stamp, tt.bp.Stamp)
				}
				if !bytes.Equal(newBp.LeaderSig, tt.bp.LeaderSig) {
					t.Errorf("LeaderSig mismatch: got %v, want %v", newBp.LeaderSig, tt.bp.LeaderSig)
				}
			}
		})
	}
}

func TestBlockProp_UnmarshalInvalidData(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "nil data",
			data:    nil,
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "insufficient data",
			data:    bytes.Repeat([]byte{1}, 8+2*types.HashLen+8+3),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &blockProp{}
			err := bp.UnmarshalBinary(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockProp_UnmarshalFromReader(t *testing.T) {
	tests := []struct {
		name    string
		reader  io.Reader
		wantErr bool
	}{
		{
			name:    "empty reader",
			reader:  bytes.NewReader([]byte{}),
			wantErr: true,
		},
		{
			name: "valid data",
			reader: bytes.NewReader(func() []byte {
				buf := make([]byte, 8+2*types.HashLen+8+4)
				binary.LittleEndian.PutUint64(buf[:8], 100)
				return buf
			}()),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &blockProp{}
			err := bp.UnmarshalFromReader(tt.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalFromReader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
