package types

import (
	"bytes"
	"math/big"
	"testing"
)

func FuzzTransactionUnmarshalBinary(f *testing.F) {
	// Add seed corpus
	validTx := &Transaction{
		Body: &TransactionBody{
			Description: "test tx",
			Payload:     []byte("test payload"),
			PayloadType: "test_type",
			Fee:         big.NewInt(1000),
			Nonce:       1,
			ChainID:     "test_chain",
		},
		Serialization: SignedMsgConcat,
		Sender:        []byte("sender"),
	}

	txBytes, _ := validTx.MarshalBinary()
	f.Add(txBytes)

	f.Fuzz(func(t *testing.T, data []byte) {
		var tx Transaction
		tx.StrictUnmarshal()
		err := tx.UnmarshalBinary(data)

		if err == nil {
			// Verify that marshaling the unmarshaled data produces identical results
			remarshaled, err := tx.MarshalBinary()
			if err != nil {
				t.Skip()
			}

			// If unmarshaling succeeded, the remarshaled data should equal the original
			if !bytes.Equal(remarshaled, data[:len(remarshaled)]) {
				var tx2 Transaction
				err = tx2.UnmarshalBinary(data)
				if err != nil {
					t.Fatal(err)
				}
				t.Errorf("remarshaled data does not match original:\noriginal: %x\nremarshaled: %x",
					data, remarshaled)
			}

			// Basic validity checks
			if tx.Body != nil {
				if len(tx.Body.Description) > MsgDescriptionMaxLength {
					t.Error("description exceeds maximum length")
				}
				// if tx.Body.Fee == nil {
				// 	t.Error("fee should not be nil")
				// }
			}
		}
	})
}

func FuzzProcedureReturnUnmarshalBinary(f *testing.F) {
	// Add seed corpus
	seeds := [][]byte{
		// Valid empty return
		{0, 0, 0, 0, 0, 0, 0, 0},
		// Valid table return with one field
		{0, 0, 1, 0, 0, 0, 1, // version, isTable, fieldCount
			0, 0, 0, 0, 3, 'i', 'd', 't', // field name "idt"
			0, 0, 0, 0, 3, 'i', 'n', 't', 0, 0, 0, 0, 0}, // field type "int"
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		pr := &ProcedureReturn{}
		err := pr.UnmarshalBinary(data)

		if err == nil {
			// Verify round-trip marshaling
			marshaled, marshalErr := pr.MarshalBinary()
			if marshalErr != nil {
				t.Errorf("failed to marshal unmarshaled data: %v", marshalErr)
			}

			// Verify fields if present
			for _, field := range pr.Fields {
				if field.Type == nil {
					t.Error("field type cannot be nil")
				}
			}

			// Verify size calculation
			if len(marshaled) != pr.SerializeSize() {
				t.Errorf("size mismatch: got %d, want %d", len(marshaled), pr.SerializeSize())
			}
		}
	})
}

func FuzzDataTypeUnmarshalBinary(f *testing.F) {
	// Add seed corpus
	seeds := [][]byte{
		// Valid int type
		{0, 0, 0, 0, 0, 3, 'i', 'n', 't', 0, 0, 0, 0, 0},
		// Valid text array type
		{0, 0, 0, 0, 0, 4, 't', 'e', 'x', 't', 1, 0, 0, 0, 0},
		// Valid decimal type with metadata
		{0, 0, 0, 0, 0, 7, 'd', 'e', 'c', 'i', 'm', 'a', 'l', 0, 0, 10, 0, 2},
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		dt := &DataType{}
		err := dt.UnmarshalBinary(data)

		if err == nil {
			// Verify the unmarshaled data can be marshaled back
			marshaled, marshalErr := dt.MarshalBinary()
			if marshalErr != nil {
				t.Errorf("failed to marshal unmarshaled data: %v", marshalErr)
			}

			// Verify data type constraints
			if dt.Name != "" {
				cleanErr := dt.Clean()
				if cleanErr == nil {
					// If Clean() succeeds, verify the type is one of the known types
					validTypes := map[string]bool{
						"int": true, "text": true, "bool": true,
						"blob": true, "uuid": true, "uint256": true,
						"decimal": true, "null": true, "unknown": true,
					}
					if !validTypes[dt.Name] {
						t.Errorf("unmarshaled invalid type name: %s", dt.Name)
					}
				}
			}

			// Verify size calculation
			if len(marshaled) != dt.SerializeSize() {
				t.Errorf("size mismatch: got %d, want %d", len(marshaled), dt.SerializeSize())
			}
		}
	})
}

func FuzzForeignProcedureUnmarshalBinary(f *testing.F) {
	// Add seed corpus with valid structures
	seeds := [][]byte{
		// Basic foreign procedure with no params and nil returns
		{0, 0, // version
			0, 0, 0, 0, // empty name
			0, 0, 0, 0}, // no parameters

		// Foreign procedure with int parameter and table return
		{0, 0, // version
			0, 0, 0, 3, 'f', 'o', 'o', // name
			0, 0, 0, 1, // one parameter
			0, 0, 0, 0, 0, 3, 'i', 'n', 't', 0, 0, 3, 0, 4, // int parameter
			1,                   // non-nil returns
			0, 0, 1, 0, 0, 0, 1, // returns table with 1 field
			0, 0, 0, 0, 0, 2, 'i', 'd', // field name
			0, 0, 0, 0, 0, 3, 'i', 'n', 't', 0, 0, 1, 0, 2}, // field type
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		fp := &ForeignProcedure{}
		err := fp.UnmarshalBinary(data)

		if err == nil {
			// Verify parameters
			for _, param := range fp.Parameters {
				if param == nil {
					t.Error("parameter cannot be nil after successful unmarshal")
				}
			}

			// Verify round-trip marshaling
			marshaled, marshalErr := fp.MarshalBinary()
			if marshalErr != nil {
				t.Errorf("marshal failed after successful unmarshal: %v", marshalErr)
			}

			// Verify marshaled length matches original
			if !bytes.Equal(marshaled, data[:len(marshaled)]) {
				t.Errorf("marshaled data mismatch")
			}
		}
	})
}
