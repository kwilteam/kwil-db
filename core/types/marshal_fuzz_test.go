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
				t.Errorf("remarshaled data does not match source:\noriginal: %x\nremarshaled: %x",
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

func FuzzDataTypeUnmarshalBinary(f *testing.F) {
	// Add seed corpus
	seeds := [][]byte{
		// Valid int type
		{0, 0, 0, 0, 0, 3, 'i', 'n', 't', '8', 0, 0, 0, 0, 0},
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
						"int8": true, "text": true, "bool": true,
						"bytea": true, "uuid": true, "uint256": true,
						"numeric": true, "null": true, "unknown": true,
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
