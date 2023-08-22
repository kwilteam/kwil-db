package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"

	ethAccount "github.com/ethereum/go-ethereum/accounts"
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

// SignMsg signs the given message(not hashed) according to EIP-191 personal_sign.
// This is default signature type for sec256k1.
// Implements the Signer interface.
func (pv *Secp256k1PrivateKey) SignMsg(msg []byte) (*Signature, error) {
	hash := ethAccount.TextHash(msg)
	sig, err := pv.Sign(hash)
	if err != nil {
		return nil, err
	}
	return &Signature{
		Signature: sig,
		Type:      SignatureTypeSecp256k1Personal,
	}, nil
}

// Sign signs the given hash directly utilizing go-ethereum's Sign function.
func (pv *Secp256k1PrivateKey) Sign(hash []byte) ([]byte, error) {
	return ethCrypto.Sign(hash, pv.key)
}

func (pv *Secp256k1PrivateKey) Signer() Signer {
	return pv
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

func (pub *Secp256k1PublicKey) Type() KeyType {
	return Secp256k1
}

// Verify verifies the signature against the given hash.
// e.g. this verify able to verify multi-signature-schema like personal_sign, eip712, cometbft, etc.
func (pub *Secp256k1PublicKey) Verify(sig []byte, hash []byte) error {
	if len(sig) != 64 {
		return errInvalidSignature
	}

	// signature should have the 64 byte [R || S] format
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
