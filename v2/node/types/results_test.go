package types

import (
	"encoding/binary"
	"testing"
)

func TestTxResultMarshalUnmarshal(t *testing.T) {
	t.Run("empty result", func(t *testing.T) {
		tr := TxResult{
			Code:   0,
			Log:    "",
			Events: nil,
		}

		data, err := tr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded TxResult
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Code != tr.Code {
			t.Errorf("got code %d, want %d", decoded.Code, tr.Code)
		}
		if decoded.Log != tr.Log {
			t.Errorf("got log %s, want %s", decoded.Log, tr.Log)
		}
		if len(decoded.Events) != 0 {
			t.Errorf("got %d events, want 0", len(decoded.Events))
		}
	})

	t.Run("with log and code", func(t *testing.T) {
		tr := TxResult{
			Code:   123,
			Log:    "test log message",
			Events: nil,
		}

		data, err := tr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded TxResult
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Code != tr.Code {
			t.Errorf("got code %d, want %d", decoded.Code, tr.Code)
		}
		if decoded.Log != tr.Log {
			t.Errorf("got log %s, want %s", decoded.Log, tr.Log)
		}
	})

	t.Run("invalid data length", func(t *testing.T) {
		data := make([]byte, 3)
		var tr TxResult
		err := tr.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for insufficient data")
		}
	})

	t.Run("invalid log length", func(t *testing.T) {
		data := make([]byte, 6)
		binary.BigEndian.PutUint16(data, uint16(1))
		binary.BigEndian.PutUint32(data[2:], uint32(1000000))

		var tr TxResult
		err := tr.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid log length")
		}
	})

	t.Run("invalid events length", func(t *testing.T) {
		tr := TxResult{
			Code:   1,
			Log:    "test",
			Events: make([]Event, 65536),
		}

		_, err := tr.MarshalBinary()
		if err == nil {
			t.Error("expected error for too many events")
		}
	})
}
