package node

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"kwil/node/types"
)

func TestBlockAnnMsg_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		msg     *blockAnnMsg
		wantErr bool
	}{
		{
			name:    "empty message",
			msg:     &blockAnnMsg{},
			wantErr: false,
		},
		{
			name: "message with data",
			msg: &blockAnnMsg{
				Height:    100,
				Hash:      [32]byte{1, 2, 3},
				AppHash:   [32]byte{4, 5, 6},
				LeaderSig: []byte{7, 8, 9},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.msg.MarshalBinary()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			newMsg := &blockAnnMsg{}
			err = newMsg.UnmarshalBinary(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if newMsg.Height != tt.msg.Height {
					t.Errorf("Height mismatch: got %v, want %v", newMsg.Height, tt.msg.Height)
				}
				if newMsg.Hash != tt.msg.Hash {
					t.Errorf("Hash mismatch: got %v, want %v", newMsg.Hash, tt.msg.Hash)
				}
			}
		})
	}
}

func TestBlockAnnMsg_UnmarshalInvalidData(t *testing.T) {
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
			name:    "invalid data",
			data:    []byte{1, 2, 3},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &blockAnnMsg{}
			err := msg.UnmarshalBinary(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestBlockHeightReq_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		req     *blockHeightReq
		wantErr bool
	}{
		{
			name:    "zero height",
			req:     &blockHeightReq{Height: 0},
			wantErr: false,
		},
		{
			name:    "positive height",
			req:     &blockHeightReq{Height: 12345},
			wantErr: false,
		},
		{
			name:    "negative height",
			req:     &blockHeightReq{Height: -12345},
			wantErr: false,
		},
		{
			name:    "max height",
			req:     &blockHeightReq{Height: math.MaxInt64},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.req.MarshalBinary()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			newReq := &blockHeightReq{}
			err = newReq.UnmarshalBinary(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && newReq.Height != tt.req.Height {
				t.Errorf("Height mismatch: got %v, want %v", newReq.Height, tt.req.Height)
			}
		})
	}
}

func TestBlockHeightReq_ReadWriteTo(t *testing.T) {
	tests := []struct {
		name    string
		req     *blockHeightReq
		wantErr bool
	}{
		{
			name:    "valid height",
			req:     &blockHeightReq{Height: 54321},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			n, err := tt.req.WriteTo(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if n != 8 {
				t.Errorf("WriteTo() wrote %d bytes, want 8", n)
			}

			newReq := &blockHeightReq{}
			n, err = newReq.ReadFrom(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if n != 8 {
				t.Errorf("ReadFrom() read %d bytes, want 8", n)
			}

			if !tt.wantErr && newReq.Height != tt.req.Height {
				t.Errorf("Height mismatch: got %v, want %v", newReq.Height, tt.req.Height)
			}
		})
	}
}

func TestBlockHashReq_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		req     *blockHashReq
		wantErr bool
	}{
		{
			name:    "zero hash",
			req:     &blockHashReq{Hash: types.Hash{}},
			wantErr: false,
		},
		{
			name: "non-zero hash",
			req: &blockHashReq{Hash: types.Hash{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
				11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
				21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.req.MarshalBinary()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			newReq := &blockHashReq{}
			err = newReq.UnmarshalBinary(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && newReq.Hash != tt.req.Hash {
				t.Errorf("Hash mismatch: got %v, want %v", newReq.Hash, tt.req.Hash)
			}
		})
	}
}

func TestBlockHashReq_ReadWriteTo(t *testing.T) {
	testHash := types.Hash{1, 2, 3, 4, 5}
	tests := []struct {
		name    string
		req     *blockHashReq
		wantErr bool
	}{
		{
			name:    "valid hash",
			req:     &blockHashReq{Hash: testHash},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			n, err := tt.req.WriteTo(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if n != types.HashLen {
				t.Errorf("WriteTo() wrote %d bytes, want %d", n, types.HashLen)
			}

			newReq := &blockHashReq{}
			n, err = newReq.ReadFrom(buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if n != types.HashLen {
				t.Errorf("ReadFrom() read %d bytes, want %d", n, types.HashLen)
			}

			if !tt.wantErr && newReq.Hash != tt.req.Hash {
				t.Errorf("Hash mismatch: got %v, want %v", newReq.Hash, tt.req.Hash)
			}
		})
	}
}

func TestBlockHeightReq_UnmarshalInvalidData(t *testing.T) {
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
			name:    "short data",
			data:    []byte{1, 2, 3},
			wantErr: true,
		},
		{
			name:    "long data",
			data:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &blockHeightReq{}
			err := req.UnmarshalBinary(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockHashReq_UnmarshalInvalidData(t *testing.T) {
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
			name:    "short data",
			data:    []byte{1, 2, 3},
			wantErr: true,
		},
		{
			name:    "long data",
			data:    bytes.Repeat([]byte{1}, types.HashLen+1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &blockHashReq{}
			err := req.UnmarshalBinary(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReadResp(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		limit   int64
		want    []byte
		wantErr error
	}{
		{
			name:    "valid response",
			input:   []byte("valid response data"),
			limit:   100,
			want:    []byte("valid response data"),
			wantErr: nil,
		},
		{
			name:    "empty response",
			input:   []byte{},
			limit:   100,
			want:    nil,
			wantErr: ErrNoResponse,
		},
		{
			name:    "noData response",
			input:   []byte{0},
			limit:   100,
			want:    nil,
			wantErr: ErrNotFound,
		},
		{
			name:    "response exceeds limit",
			input:   bytes.Repeat([]byte("a"), 1000),
			limit:   10,
			want:    bytes.Repeat([]byte("a"), 10),
			wantErr: nil,
		},
		{
			name:    "zero limit",
			input:   []byte("test data"),
			limit:   0,
			want:    nil,
			wantErr: ErrNoResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			got, err := readResp(reader, tt.limit)

			if err != tt.wantErr {
				t.Errorf("readResp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil && !bytes.Equal(got, tt.want) {
				t.Errorf("readResp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadResp_BrokenReader(t *testing.T) {
	errTest := errors.New("read error")
	brokenReader := &brokenReader{err: errTest}

	_, err := readResp(brokenReader, 100)
	if err != errTest {
		t.Errorf("readResp() error = %v, want %v", err, errTest)
	}
}

type brokenReader struct {
	err error
}

func (br *brokenReader) Read(p []byte) (n int, err error) {
	return 0, br.err
}
