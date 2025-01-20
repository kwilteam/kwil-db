//go:build rsa2048_kwil_key || ext_test

package crypto

import (
	stdCrypto "crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
)

const KeyTypeRSA2048 KeyType = "rsa2048"

type RSA2048Definition struct{}

var _ KeyDefinition = RSA2048Definition{}

func init() {
	err := RegisterKeyType(RSA2048Definition{})
	if err != nil {
		panic(err)
	}
}

func (RSA2048Definition) Type() KeyType {
	return KeyTypeRSA2048
}

func (RSA2048Definition) EncodeFlag() uint32 {
	return ReservedKeyIDs + 1
}

func (RSA2048Definition) UnmarshalPrivateKey(b []byte) (PrivateKey, error) {
	return UnmarshalRSAPrivateKey(b)
}

func (RSA2048Definition) UnmarshalPublicKey(b []byte) (PublicKey, error) {
	return UnmarshalRSAPublicKey(b)
}

func (RSA2048Definition) Generate() PrivateKey {
	priv, _, _ := GenerateRSAKey(nil, 2048)
	return priv
}

func GenerateRSAKey(src io.Reader, bits int) (*RSAPrivateKey, *RSAPublicKey, error) {
	if src == nil {
		src = rand.Reader
	}
	priv, err := rsa.GenerateKey(src, bits)
	if err != nil {
		return nil, nil, err
	}
	return &RSAPrivateKey{privKey: priv}, &RSAPublicKey{pubKey: &priv.PublicKey}, nil
}

type RSAPublicKey struct {
	pubKey *rsa.PublicKey
}

var _ Key = &RSAPublicKey{}

func (rpub *RSAPublicKey) Type() KeyType {
	return KeyTypeRSA2048
}

func (rpub *RSAPublicKey) Bytes() []byte {
	b := binary.LittleEndian.AppendUint64(nil, uint64(rpub.pubKey.E))
	return append(b, rpub.pubKey.N.Bytes()...)
}

func (rpub *RSAPublicKey) Equals(k Key) bool {
	other, ok := k.(*RSAPublicKey)
	if !ok {
		return false
	}
	// N and E
	return rpub.pubKey.N.Cmp(other.pubKey.N) == 0 && rpub.pubKey.E == other.pubKey.E
}

var _ PublicKey = &RSAPublicKey{}

func (rpub *RSAPublicKey) Verify(data []byte, sig []byte) (bool, error) {
	sigHash := sha256.Sum256(data)
	err := rsa.VerifyPKCS1v15(rpub.pubKey, stdCrypto.SHA256, sigHash[:], sig)
	if err != nil {
		return false, err
	}
	return true, nil
}

// UnmarshalRSAPrivateKey matches (*RSAPrivateKey).Bytes()
func UnmarshalRSAPrivateKey(b []byte) (PrivateKey, error) {
	priv, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		return nil, err
	}
	// check bitlen of N, should be between 2048 and 8192
	bits := priv.PublicKey.N.BitLen()
	if bits < 2048 || bits > 8192 {
		return nil, fmt.Errorf("invalid RSA private key")
	}
	return &RSAPrivateKey{privKey: priv}, nil
}

// UnmarshalRSAPublicKey matches (*RSAPublicKey).Bytes()
func UnmarshalRSAPublicKey(b []byte) (PublicKey, error) {
	if len(b) < 16 {
		return nil, fmt.Errorf("invalid RSA private key")
	}

	priv := &rsa.PublicKey{}
	priv.E = int(binary.LittleEndian.Uint64(b[:8]))
	priv.N = new(big.Int).SetBytes(b[8:])

	// check bitlen of N, should be between 2048 and 8192
	bits := priv.N.BitLen()
	if bits < 2048 || bits > 8192 {
		return nil, fmt.Errorf("invalid RSA private key")
	}

	return &RSAPublicKey{pubKey: priv}, nil
}

type RSAPrivateKey struct {
	privKey *rsa.PrivateKey
}

var _ Key = &RSAPrivateKey{}

func (r *RSAPrivateKey) Type() KeyType {
	return KeyTypeRSA2048
}

func (r *RSAPrivateKey) Bytes() []byte {
	return x509.MarshalPKCS1PrivateKey(r.privKey)
}

func (r *RSAPrivateKey) ASCII() (rep, data string) {
	// bts, _ := x509.MarshalPKCS8PrivateKey(r.privKey)
	return "PKCS #1, ASN.1 DER base64", "\n-----BEGIN RSA PRIVATE KEY-----\n" +
		base64.StdEncoding.EncodeToString(r.Bytes()) + "\n-----END RSA PRIVATE KEY-----"
}

func (r *RSAPrivateKey) Equals(k Key) bool {
	other, ok := k.(*RSAPrivateKey)
	if !ok {
		return r.Type() != k.Type() && subtle.ConstantTimeCompare(r.Bytes(), k.Bytes()) == 1
	}
	return r.privKey.Equal(other.privKey)

	// N and E
	// return r.privKey.N.Cmp(other.privKey.N) == 0 && r.privKey.E == other.privKey.E
}

var _ PrivateKey = (*RSAPrivateKey)(nil)

func (r *RSAPrivateKey) Public() PublicKey {
	return &RSAPublicKey{pubKey: &r.privKey.PublicKey}
}

// Sign
func (r *RSAPrivateKey) Sign(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	// r.privKey.Sign()
	return rsa.SignPKCS1v15(nil, r.privKey, stdCrypto.SHA256, hash[:])
}
