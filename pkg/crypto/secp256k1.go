package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

const Secp256k1 KeyType = "secp256k1"

type Secp256k1PrivateKey struct {
	key *ecdsa.PrivateKey
}

func (pv *Secp256k1PrivateKey) Bytes() []byte {
	return ethCrypto.FromECDSA(pv.key)
}

func (pv *Secp256k1PrivateKey) Hex() string {
	return hex.EncodeToString(pv.Bytes())
}

func (pv *Secp256k1PrivateKey) PubKey() PublicKey {
	return &Secp256k1PublicKey{
		publicKey: &pv.key.PublicKey,
	}
}

// Sign signs the given hash directly utilizing go-ethereum's Sign function.
// This returns a standard secp256k1 signature, in [R || S || V] format where V is 0 or 1, 65 bytes long.
func (pv *Secp256k1PrivateKey) Sign(hash []byte) ([]byte, error) {
	return ethCrypto.Sign(hash, pv.key)
}

func (pv *Secp256k1PrivateKey) Type() KeyType {
	return Secp256k1
}

type Secp256k1PublicKey struct {
	publicKey *ecdsa.PublicKey
}

func (pub *Secp256k1PublicKey) Address() Address {
	return &Secp256k1Address{
		address: ethCrypto.PubkeyToAddress(*pub.publicKey),
	}
}

func (pub *Secp256k1PublicKey) Bytes() []byte {
	return ethCrypto.FromECDSAPub(pub.publicKey)
}

// CompressedBytes returns the compressed bytes of the public key.
func (pub *Secp256k1PublicKey) CompressedBytes() []byte {
	return ethCrypto.CompressPubkey(pub.publicKey)
}

func (pub *Secp256k1PublicKey) Type() KeyType {
	return Secp256k1
}

// Verify verifies the standard secp256k1 signature against the given hash.
// Caller of this function should make sure the signature is in one of the following two formats:
// - 65 bytes, [R || S || V] format. This is the standard format.
// - 64 bytes, [R || S] format.
//
// Since `Verify` suppose to verify the signature produced from `Sign` function, it expects the signature to be
// 65 bytes long, and in [R || S || V] format where V is 0 or 1.
// In this implementation, we use `VerifySignature`, which doesn't care about the recovery ID, so it can
// also support 64 bytes [R || S] format signature like cometbft.
// e.g. this `Verify` function is able to verify multi-signature-schema like personal_sign, eip712, cometbft, etc.,
// as long as the given signature is in supported format.
func (pub *Secp256k1PublicKey) Verify(sig []byte, hash []byte) error {
	if len(sig) == 65 {
		// we choose `VerifySignature` since it doesn't care recovery ID
		// it expects signature in 64 byte [R || S] format
		sig = sig[:len(sig)-1]
	}

	if len(sig) != 64 {
		return errInvalidSignature
	}

	if !ethCrypto.VerifySignature(pub.Bytes(), hash, sig) {
		return errVerifySignatureFailed
	}

	return nil
}

type Secp256k1Address struct {
	address common.Address
}

func (addr *Secp256k1Address) Bytes() []byte {
	return addr.address.Bytes()
}

func (addr *Secp256k1Address) Type() KeyType {
	return Secp256k1
}

func (addr *Secp256k1Address) String() string {
	return addr.address.Hex()
}

// GenerateSecp256k1Key generates a new secp256k1 private key.
func GenerateSecp256k1Key() (*Secp256k1PrivateKey, error) {
	key, err := ethCrypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return &Secp256k1PrivateKey{
		key: key,
	}, nil
}
