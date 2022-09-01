package crypto

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func (p *PrivateKey) sign(data []byte) (string, error) {
	return Sign(data, p.key)
}

// I (brennan) need to correct this area of code.  We need this sign function externally but the pkey.sign function is also a dependency
func Sign(data []byte, k *ecdsa.PrivateKey) (string, error) {
	hash := crypto.Keccak256Hash(data)
	sig, err := crypto.Sign(hash.Bytes(), k)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(sig), nil
}

func ECDSAFromHex(hex string) (*ecdsa.PrivateKey, error) {
	return crypto.HexToECDSA(hex)
}

func CheckSignature(addr, sig string, data []byte) (bool, error) {
	hash := crypto.Keccak256Hash(data)
	sb, err := hexutil.Decode(sig)
	if err != nil {
		return false, err
	}

	pubBytes, err := crypto.Ecrecover(hash.Bytes(), sb)
	if err != nil {
		return false, err // I don't believe this can be reached
	}

	pub, err := crypto.UnmarshalPubkey(pubBytes)
	if err != nil {
		return false, err // I don't believe this can be reached
	}
	derAddr := crypto.PubkeyToAddress(*pub)

	return derAddr == common.HexToAddress(addr), nil
}
