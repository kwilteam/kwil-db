package crypto

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ec "github.com/ethereum/go-ethereum/crypto"
)

func (p *PrivateKey) sign(data []byte) (string, error) {
	return Sign(data, p.key)
}

// I (brennan) need to correct this area of code.  We need this sign function externally but the pkey.sign function is also required for the tests
func Sign(data []byte, k *ecdsa.PrivateKey) (string, error) {
	hash := ec.Keccak256Hash(data)
	sig, err := ec.Sign(hash.Bytes(), k)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(sig), nil
}

func ECDSAFromHex(hex string) (*ecdsa.PrivateKey, error) {
	return ec.HexToECDSA(hex)
}

func CheckSignature(addr, sig string, data []byte) (bool, error) {
	hash := ec.Keccak256Hash(data)
	sb, err := hexutil.Decode(sig)
	if err != nil {
		return false, err
	}

	pubBytes, err := ec.Ecrecover(hash.Bytes(), sb)
	if err != nil {
		return false, err // I don't believe this can be reached
	}

	pub, err := ec.UnmarshalPubkey(pubBytes)
	if err != nil {
		return false, err // I don't believe this can be reached
	}
	derAddr := ec.PubkeyToAddress(*pub)

	return derAddr == common.HexToAddress(addr), nil
}
