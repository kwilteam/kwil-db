package types

import (
	"bytes"
	"testing"
)

func TestConsensusReset_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		height  int64
		wantErr bool
	}{
		{
			name:    "positive height",
			height:  100,
			wantErr: false,
		},
		{
			name:    "zero height",
			height:  0,
			wantErr: false,
		},
		{
			name:    "negative height",
			height:  -1,
			wantErr: false,
		},
		{
			name:    "max height",
			height:  1<<63 - 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := ConsensusReset{ToHeight: tt.height}

			data, err := cr.MarshalBinary()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var decoded ConsensusReset
			err = decoded.UnmarshalBinary(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && decoded.ToHeight != tt.height {
				t.Errorf("Round trip failed: got %v, want %v", decoded.ToHeight, tt.height)
			}
		})
	}
}

func TestConsensusReset_UnmarshalInvalid(t *testing.T) {
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
			data:    []byte{1, 2, 3},
			wantErr: true,
		},
		{
			name:    "excess data",
			data:    bytes.Repeat([]byte{1}, 9),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cr ConsensusReset
			err := cr.UnmarshalBinary(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConsensusReset_Bytes(t *testing.T) {
	tests := []struct {
		name   string
		height int64
		want   int
	}{
		{
			name:   "standard height",
			height: 1000,
			want:   8,
		},
		{
			name:   "zero height",
			height: 0,
			want:   8,
		},
		{
			name:   "max height",
			height: 1<<63 - 1,
			want:   8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := ConsensusReset{ToHeight: tt.height}
			got := cr.Bytes()
			if len(got) != tt.want {
				t.Errorf("Bytes() length = %v, want %v", len(got), tt.want)
			}
		})
	}
}
