package crypto

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type KeyDefinition interface {
	Type() KeyType      // name of key type e.g. "secp256k1"
	EncodeFlag() uint32 // prefix for compact unique binary encoding

	UnmarshalPrivateKey(b []byte) (PrivateKey, error)
	UnmarshalPublicKey(b []byte) (PublicKey, error)

	Generate() PrivateKey
}

const (
	// ReservedKeyIDs is the range of key types that are reserved for internal purposes.
	ReservedKeyIDs = 1 << 16 // 65536
)

var (
	// keyTypes maps key types to their string representations
	// of all the supported key types.
	keyTypes = map[KeyType]KeyDefinition{
		KeyTypeSecp256k1: Secp256k1Definition{},
		KeyTypeEd25519:   Ed25519Definition{},
	}

	encodingIDs = map[uint32]KeyType{
		Secp256k1Definition{}.EncodeFlag(): KeyTypeSecp256k1,
		Ed25519Definition{}.EncodeFlag():   KeyTypeEd25519,
	}
)

// RegisterKeyType registers a new keyType. The KeyType and its string should be
// unique and the KeyType value must be greater than ReservedKeyTypes.
func RegisterKeyType(kd KeyDefinition) error {
	kt := kd.Type()
	encID := kd.EncodeFlag()
	if encID <= ReservedKeyIDs {
		return fmt.Errorf("key type %s (%d) is a reserved ID", kt, encID)
	}

	if strings.ContainsRune(kt.String(), ' ') {
		return fmt.Errorf("key type string %s contains spaces", kt)
	}

	if _, ok := keyTypes[kt]; ok {
		return fmt.Errorf("key type %s already registered", kt)
	}

	if kt0, ok := encodingIDs[encID]; ok {
		return fmt.Errorf("key encoding prefix %d already registered with keyType: %s", encID, kt0)
	}

	keyTypes[kt] = kd
	encodingIDs[encID] = kt

	return nil
}

func KeyTypeDefinition(kt KeyType) (KeyDefinition, bool) {
	kd, ok := keyTypes[kt]
	return kd, ok
}

// ParseKeyType parses a string into a KeyType. This ensures that the string has
// a registered key definition.
func ParseKeyType(s string) (KeyType, error) {
	if kd, ok := KeyTypeDefinition(KeyType(s)); ok {
		return kd.Type(), nil
	}
	return "", fmt.Errorf("unknown key type: %s", s)
}

func UnmarshalPublicKey(b []byte, kt KeyType) (PublicKey, error) {
	kd, ok := keyTypes[kt]
	if !ok {
		return nil, fmt.Errorf("unknown key type: %v", kt)
	}
	return kd.UnmarshalPublicKey(b)
}

func UnmarshalPrivateKey(b []byte, kt KeyType) (PrivateKey, error) {
	kd, ok := keyTypes[kt]
	if !ok {
		return nil, fmt.Errorf("unknown key type: %v", kt)
	}
	return kd.UnmarshalPrivateKey(b)
}

func GeneratePrivateKey(kt KeyType) (PrivateKey, error) {
	kd, ok := keyTypes[kt]
	if !ok {
		return nil, fmt.Errorf("unknown key type: %v", kt)
	}
	return kd.Generate(), nil
}

func WireEncodeKeyType(kt KeyType) []byte {
	kd, ok := keyTypes[kt]
	if !ok {
		panic("unknown key type")
	}
	return binary.LittleEndian.AppendUint32(nil, kd.EncodeFlag())
}

func WireDecodeKeyType(b []byte) (KeyType, []byte, error) {
	if len(b) < 4 {
		return "", nil, fmt.Errorf("invalid key type encoding")
	}
	encID := binary.LittleEndian.Uint32(b)
	kt, ok := encodingIDs[encID]
	if !ok {
		return "", nil, fmt.Errorf("unknown key type encoding: %d", encID)
	}
	return kt, b[4:], nil
}

func WireEncodeKey(key Key) []byte {
	kd, ok := keyTypes[key.Type()]
	if !ok {
		panic("unknown key type")
	}
	b := binary.LittleEndian.AppendUint32(nil, kd.EncodeFlag())
	return append(b, key.Bytes()...)
	// buf := &bytes.Buffer{}
	// types.WriteString(buf, string(key.Type()))
	// buf.Write(key.Bytes())
	// return buf.Bytes()
}

func WireDecodePubKey(b []byte) (PublicKey, error) {
	kt, b, err := WireDecodeKeyType(b)
	if err != nil {
		return nil, err
	}
	return UnmarshalPublicKey(b, kt)
}

func WireDecodePrivateKey(b []byte) (PrivateKey, error) {
	kt, b, err := WireDecodeKeyType(b)
	if err != nil {
		return nil, err
	}
	return UnmarshalPrivateKey(b, kt)
}
