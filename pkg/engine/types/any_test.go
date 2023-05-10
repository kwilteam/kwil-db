package types_test

import (
	"bytes"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"testing"
)

func Test_Any(t *testing.T) {
	// test null
	null, err := types.New(nil)
	if err != nil {
		t.Errorf("failed to create null any: %v", err)
	}

	if null.Type() != types.NULL {
		t.Errorf("expected %d, got %d", types.NULL, null.Type())
	}

	null2, err := types.NewFromSerial(nil)
	if err != nil {
		t.Errorf("failed to create null any from serial: %v", err)
	}

	if null2.Value() != null.Value() {
		t.Errorf("expected %v, got %v", null.Value(), null2.Value())
	}

	// test int
	int2, err := types.New(int(100))
	if err != nil {
		t.Errorf("failed to create int any: %v", err)
	}

	if int2.Type() != types.INT {
		t.Errorf("expected %d, got %d", types.INT, int2.Type())
	}

	val := int2.Value()

	if val.(int) != 100 {
		t.Errorf("expected %v, got %v", 100, val)
	}

	// get int bytes
	bts := int2.Bytes()

	// try to create a new any from the bytes
	int3, err := types.NewFromSerial(bts)
	if err != nil {
		t.Errorf("failed to create new any from serial: %v", err)
	}

	if int3.Type() != types.INT {
		t.Errorf("expected %d, got %d", types.INT, int3.Type())
	}

	val = int3.Value()

	if val.(int64) != 100 {
		t.Errorf("expected %v, got %v", 100, val.(int64))
	}

	val = int3.Value()

	if val.(int64) != 100 {
		t.Errorf("expected %v, got %v", 100, val)
	}

	// re-serialize
	bts2 := int3.Bytes()

	if !bytes.Equal(bts, bts2) {
		t.Errorf("expected %v, got %v", bts, bts2)
	}
}
