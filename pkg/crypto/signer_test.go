package crypto

import (
	"encoding/hex"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComebftSecp256k1Signer_SignMsg(t *testing.T) {
	msg := []byte("foo")

	pvKeyHex := "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"
	sigHex := "19a4aced02d5b9142b4f622b06442b1904445e16bd25409e6b0ff357ccc021d001d0e7824654b695b4b6e0991cb7507f487b82be4b2ed713d1e3e2cbc3d2518d"

	pvKeyBytes, _ := hex.DecodeString(pvKeyHex)
	cometBftSecp256k1Key := secp256k1.PrivKey(pvKeyBytes)
	cometBfgSecp256k1Sig, err := cometBftSecp256k1Key.Sign(msg)
	assert.NoError(t, err, "error signing message")
	assert.Equal(t, sigHex, hex.EncodeToString(cometBfgSecp256k1Sig), "signature mismatch")

	// use the cometbft signer to sign the message
	kwilCometBftKey, _ := Secp256k1PrivateKeyFromHex(pvKeyHex)
	cometBfgSigner := &ComebftSecp256k1Signer{
		key: kwilCometBftKey,
	}
	kwilCometBftKeySig, err := cometBfgSigner.SignMsg(msg)
	assert.NoError(t, err, "error signing message")
	assert.Equal(t, sigHex, hex.EncodeToString(kwilCometBftKeySig.Signature), "signature mismatch")
}
