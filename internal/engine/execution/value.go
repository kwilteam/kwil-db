package execution

import "math/big"

// a Value is a text, int, boolean, uuid, text[], composite type, etc.
type Value interface{}

type TextValue string
type IntValue int
type Uint256Value *big.Int
type BooleanValue bool
type UUIDValue string
type BlobValue []byte
type ArrayValue []Value
type CompositeValue map[string]Value
