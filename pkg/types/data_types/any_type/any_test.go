package anytype_test

import (
	"bytes"
	"kwil/pkg/types/data_types"
	"kwil/pkg/types/data_types/any_type"
	"testing"
)

func Test_Any(t *testing.T) {
	// test null
	null, err := anytype.New(nil)
	if err != nil {
		t.Errorf("failed to create null any: %v", err)
	}

	if null.Type() != datatypes.NULL {
		t.Errorf("expected %d, got %d", datatypes.NULL, null.Type())
	}

	null2, err := anytype.NewFromSerial(nil)
	if err != nil {
		t.Errorf("failed to create null any from serial: %v", err)
	}

	if null2.Value() != null.Value() {
		t.Errorf("expected %v, got %v", null.Value(), null2.Value())
	}

	// test bool
	bool1, err := anytype.New(true)
	if err != nil {
		t.Errorf("failed to create bool any: %v", err)
	}

	if bool1.Type() != datatypes.BOOLEAN {
		t.Errorf("expected %d, got %d", datatypes.BOOLEAN, bool1.Type())
	}

	val := bool1.Value()

	if val != true {
		t.Errorf("expected %v, got %v", true, val)
	}

	// test int32
	int1, err := anytype.New(int32(100))
	if err != nil {
		t.Errorf("failed to create int any: %v", err)
	}

	if int1.Type() != datatypes.INT32 {
		t.Errorf("expected %d, got %d", datatypes.INT32, int1.Type())
	}

	val = int1.Value()

	if val.(int32) != 100 {
		t.Errorf("expected %v, got %v", 100, val)
	}

	// test int64
	int2, err := anytype.New(int64(100))
	if err != nil {
		t.Errorf("failed to create int any: %v", err)
	}

	if int2.Type() != datatypes.INT64 {
		t.Errorf("expected %d, got %d", datatypes.INT64, int2.Type())
	}

	val = int2.Value()

	if val.(int64) != 100 {
		t.Errorf("expected %v, got %v", 100, val)
	}

	// get int64 bytes
	bts := int2.Bytes()

	// try to create a new any from the bytes
	int3, err := anytype.NewFromSerial(bts)
	if err != nil {
		t.Errorf("failed to create new any from serial: %v", err)
	}

	if int3.Type() != datatypes.INT64 {
		t.Errorf("expected %d, got %d", datatypes.INT64, int3.Type())
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
