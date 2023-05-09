package spec_test

import (
	"bytes"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
	"testing"
)

func Test_Any(t *testing.T) {
	// test null
	null, err := spec.New(nil)
	if err != nil {
		t.Errorf("failed to create null any: %v", err)
	}

	if null.Type() != spec.NULL {
		t.Errorf("expected %d, got %d", spec.NULL, null.Type())
	}

	null2, err := spec.NewFromSerial(nil)
	if err != nil {
		t.Errorf("failed to create null any from serial: %v", err)
	}

	if null2.Value() != null.Value() {
		t.Errorf("expected %v, got %v", null.Value(), null2.Value())
	}

	// test bool
	bool1, err := spec.New(true)
	if err != nil {
		t.Errorf("failed to create bool any: %v", err)
	}

	if bool1.Type() != spec.BOOLEAN {
		t.Errorf("expected %d, got %d", spec.BOOLEAN, bool1.Type())
	}

	val := bool1.Value()

	if val != true {
		t.Errorf("expected %v, got %v", true, val)
	}

	// test int32
	int1, err := spec.New(int32(100))
	if err != nil {
		t.Errorf("failed to create int any: %v", err)
	}

	if int1.Type() != spec.INT32 {
		t.Errorf("expected %d, got %d", spec.INT32, int1.Type())
	}

	val = int1.Value()

	if val.(int32) != 100 {
		t.Errorf("expected %v, got %v", 100, val)
	}

	// test int64
	int2, err := spec.New(int64(100))
	if err != nil {
		t.Errorf("failed to create int any: %v", err)
	}

	if int2.Type() != spec.INT64 {
		t.Errorf("expected %d, got %d", spec.INT64, int2.Type())
	}

	val = int2.Value()

	if val.(int64) != 100 {
		t.Errorf("expected %v, got %v", 100, val)
	}

	// get int64 bytes
	bts := int2.Bytes()

	// try to create a new any from the bytes
	int3, err := spec.NewFromSerial(bts)
	if err != nil {
		t.Errorf("failed to create new any from serial: %v", err)
	}

	if int3.Type() != spec.INT64 {
		t.Errorf("expected %d, got %d", spec.INT64, int3.Type())
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
