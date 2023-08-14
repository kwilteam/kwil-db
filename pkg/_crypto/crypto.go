package crypto

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ec "github.com/ethereum/go-ethereum/crypto"
)

func ECDSAFromHex(hex string) (*ecdsa.PrivateKey, error) {
	return ec.HexToECDSA(hex)
}

func AddressFromPrivateKey(key *ecdsa.PrivateKey) string {
	caddr := ec.PubkeyToAddress(key.PublicKey)
	return caddr.Hex()
}

func IsValidAddress(addr string) bool {
	return common.IsHexAddress(addr)
}

func HexFromECDSAPrivateKey(key *ecdsa.PrivateKey) string {
	return hexutil.Encode(ec.FromECDSA(key))[2:]
}
