package crypto

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func (p *PrivateKey) Sign(data []byte) (string, error) {
	hash := crypto.Keccak256Hash(data)
	sig, err := crypto.Sign(hash.Bytes(), p.key)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(sig), nil
}

func CheckSignature(addr string, sig string, data []byte) (bool, error) {
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
