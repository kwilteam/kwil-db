package anytype_test

import (
	"bytes"
	datatypes "kwil/x/types/data_types"
	anytype "kwil/x/types/data_types/any_type"
	"testing"
)

func Test_Any(t *testing.T) {
	// test null
	null, err := anytype.New(nil)
	if err != nil {
		t.Errorf("failed to create null any: %v", err)
	}

	if null.Type != datatypes.NULL {
		t.Errorf("expected %v, got %v", datatypes.NULL, null.Type)
	}

	// test bool
	bool1, err := anytype.New(true)
	if err != nil {
		t.Errorf("failed to create bool any: %v", err)
	}

	if bool1.Type != datatypes.BOOLEAN {
		t.Errorf("expected %v, got %v", datatypes.BOOLEAN, bool1.Type)
	}

	if bool1.Value.(bool) != true {
		t.Errorf("expected %v, got %v", true, bool1.Value.(bool))
	}

	// test int32
	int1, err := anytype.New(int32(100))
	if err != nil {
		t.Errorf("failed to create int any: %v", err)
	}

	if int1.Type != datatypes.INT32 {
		t.Errorf("expected %v, got %v", datatypes.INT32, int1.Type)
	}

	if int1.Value.(int32) != 100 {
		t.Errorf("expected %v, got %v", 100, int1.Value.(int32))
	}

	// test int64
	int2, err := anytype.New(int64(100))
	if err != nil {
		t.Errorf("failed to create int any: %v", err)
	}

	if int2.Type != datatypes.INT64 {
		t.Errorf("expected %v, got %v", datatypes.INT64, int2.Type)
	}

	if int2.Value.(int64) != 100 {
		t.Errorf("expected %v, got %v", 100, int2.Value.(int64))
	}

	// get int64 bytes
	bts, err := int2.GetSerialized()
	if err != nil {
		t.Errorf("failed to get serialized value: %v", err)
	}

	// try to create a new any from the bytes
	int3, err := anytype.NewFromSerial(bts)
	if err != nil {
		t.Errorf("failed to create new any from serial: %v", err)
	}

	if int3.Type != datatypes.INT64 {
		t.Errorf("expected %v, got %v", datatypes.INT64, int3.Type)
	}

	val, err := int3.Unserialize()
	if err != nil {
		t.Errorf("failed to get unserialized value: %v", err)
	}

	if val.(int64) != 100 {
		t.Errorf("expected %v, got %v", 100, val.(int64))
	}

	if int3.Value.(int64) != 100 {
		t.Errorf("expected %v, got %v", 100, int3.Value.(int64))
	}

	// re-serialize
	bts2, err := int3.Serialize()
	if err != nil {
		t.Errorf("failed to get serialized value: %v", err)
	}

	if !bytes.Equal(bts, bts2) {
		t.Errorf("expected %v, got %v", bts, bts2)
	}

}
