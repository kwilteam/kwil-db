package auth

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/kwilteam/kwil-db/core/crypto"
	"golang.org/x/crypto/sha3"
)

// Signature is a signature with a designated AuthType, which should
// be used to determine how to verify the signature.
// It seems a bit weird to have a field "Signature" inside a struct called "Signature",
// but I am keeping it like this for compatibility with the old code.
type Signature struct {
	// Data is the raw signature bytes
	Data []byte `json:"sig"`
	// Type is the signature type, which must have a registered Authenticator of
	// the same name for the Verify method to be usable.
	Type string `json:"type"`
}

func (s Signature) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(s.Data))); err != nil {
		return nil, fmt.Errorf("failed to write signature length: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, s.Data); err != nil {
		return nil, fmt.Errorf("failed to write signature data: %w", err)
	}

	if err := binary.Write(buf, binary.LittleEndian, uint32(len(s.Type))); err != nil {
		return nil, fmt.Errorf("failed to write signature type length: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, []byte(s.Type)); err != nil {
		return nil, fmt.Errorf("failed to write signature type: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *Signature) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	n, err := s.ReadFrom(r)
	if err != nil {
		return err
	}
	if len(data) != int(n) {
		return errors.New("extra signature data")
	}
	if r.Len() != 0 {
		return errors.New("extra signature data (reader)")
	}
	return nil
}

func (s *Signature) ReadFrom(r io.Reader) (int64, error) {
	rl, _ := r.(interface{ Len() int })
	var n int64
	var sigLen uint32
	if err := binary.Read(r, binary.LittleEndian, &sigLen); err != nil {
		return 0, fmt.Errorf("failed to read signature length: %w", err)
	}
	n += 4

	if sigLen > 0 {
		if rl != nil {
			if int(sigLen) > rl.Len() {
				return 0, fmt.Errorf("impossibly long signature length: %d", sigLen)
			}
			s.Data = make([]byte, sigLen)
			if _, err := io.ReadFull(r, s.Data); err != nil {
				return 0, fmt.Errorf("failed to read signature data: %w", err)
			}
		} else {
			sigBuf := &bytes.Buffer{}
			_, err := io.CopyN(sigBuf, r, int64(sigLen))
			if err != nil {
				return 0, fmt.Errorf("failed to read signature data: %w", err)
			}
			s.Data = sigBuf.Bytes()
		}

	}
	n += int64(sigLen)

	var typeLen uint32
	if err := binary.Read(r, binary.LittleEndian, &typeLen); err != nil {
		return 0, fmt.Errorf("failed to read signature type length: %w", err)
	}
	n += 4

	if typeLen > 0 {
		if rl != nil && int(typeLen) > rl.Len() {
			return 0, fmt.Errorf("impossibly long sig type length: %d", typeLen)
		}
		typeBytes := make([]byte, typeLen)
		if _, err := io.ReadFull(r, typeBytes); err != nil {
			return 0, fmt.Errorf("failed to read signature type: %w", err)
		}
		s.Type = string(typeBytes)
	}
	n += int64(typeLen)

	return n, nil
}

// Signer is an interface for something that can sign messages.
// It returns signatures with a designated AuthType, which should
// be used to determine how to verify the signature.
type Signer interface {
	// Sign signs a message and returns the signature
	Sign(msg []byte) (*Signature, error)

	// Identity returns the signer identity, which is typically and address or a
	// public key. This value is recognized by the Verify method of the
	// corresponding Authenticator for the types of signatures generated by this
	// Signer.
	Identity() []byte

	// AuthType is the type of Authenticator that should be used to verify the
	// signature created by this Signer; also to get the string identifier from
	// the Identity.
	AuthType() string
}

func GetSigner(key crypto.PrivateKey) Signer {
	switch key := key.(type) {
	case *crypto.Secp256k1PrivateKey:
		return &EthPersonalSigner{Key: *key}
	case *crypto.Ed25519PrivateKey:
		return &Ed25519Signer{Ed25519PrivateKey: *key}
	default:
		return nil
	}
}

// EthPersonalSecp256k1Signer is a signer that signs messages using the
// secp256k1 curve, using ethereum's personal_sign signature scheme.
type EthPersonalSigner struct {
	Key crypto.Secp256k1PrivateKey
}

var _ Signer = (*EthPersonalSigner)(nil)

func textHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(msg))
	return hasher.Sum(nil)
}

// Sign sign given message according to EIP-191 personal_sign.
// EIP-191 personal_sign prefix the message with "\x19Ethereum Signed Message:\n"
// and the message length, then hash the message with 'legacy' keccak256.
// The signature is in [R || S || V] format, 65 bytes.
// This method is used to sign an arbitrary message in the same manner in which
// a wallet like MetaMask would sign a text message. The message is defined by
// the object that is being serialized e.g. a Kwil Transaction.
func (e *EthPersonalSigner) Sign(msg []byte) (*Signature, error) {
	hash := textHash(msg)
	sigBts, err := e.Key.SignRaw(hash)
	if err != nil {
		return nil, err
	}

	return &Signature{
		Data: sigBts,
		Type: EthPersonalSignAuth,
	}, nil
}

// Identity returns the identity of the signer (ETH address for this signer).
func (e *EthPersonalSigner) Identity() []byte {
	pubKey := e.Key.Public().(*crypto.Secp256k1PublicKey)
	return crypto.EthereumAddressFromPubKey(pubKey)
}

func (e *EthPersonalSigner) AuthType() string {
	return EthPersonalSignAuth
}

// Ed25519Signer is a signer that signs messages using the
// ed25519 curve, using the standard signature scheme.
type Ed25519Signer struct {
	crypto.Ed25519PrivateKey
}

var _ Signer = (*Ed25519Signer)(nil)

// Sign signs the given message(not hashed) according to standard signature scheme.
// It does not apply any special digests to the message.
func (e *Ed25519Signer) Sign(msg []byte) (*Signature, error) {
	signatureBts, err := e.Ed25519PrivateKey.Sign(msg)
	if err != nil {
		return nil, err
	}

	return &Signature{
		Data: signatureBts,
		Type: Ed25519Auth,
	}, nil
}

// Identity returns the identity of the signer (public key for this signer).
func (e *Ed25519Signer) Identity() []byte {
	return e.Ed25519PrivateKey.Public().Bytes()
}

func (e *Ed25519Signer) AuthType() string {
	return Ed25519Auth
}
