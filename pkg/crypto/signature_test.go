package crypto_test

import (
	"bytes"
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ec "github.com/ethereum/go-ethereum/crypto"
)

func TestSign(t *testing.T) {
	pk, err := ec.HexToECDSA("4bb214b1f3a0737d758bc3828cdff371e3769fe84a2678da34700cb18d50770e")
	if err != nil {
		t.Errorf("failed to get test private key: %d", err)
	}

	sig, err := crypto.Sign([]byte("kwil"), pk)
	if err != nil {
		t.Errorf("failed to sign: %d", err)
	}

	expected, err := hexutil.Decode("0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501")
	if err != nil {
		t.Errorf("failed to decode expected signature: %d", err)
	}

	if !bytes.Equal(sig.Signature, expected) {
		t.Errorf("expected %s, got %s", expected, sig.Signature)
	}
}

func TestCheckSignature(t *testing.T) {
	pk, err := ec.HexToECDSA("4bb214b1f3a0737d758bc3828cdff371e3769fe84a2678da34700cb18d50770e")
	if err != nil {
		t.Errorf("failed to get test private key: %d", err)
	}

	msg := []byte("kwil")

	sig, err := crypto.Sign(msg, pk)
	if err != nil {
		t.Errorf("failed to sign: %d", err)
	}

	ok, err := crypto.CheckSignature("0x995d95245698212D4Af52c8031F614C3D3127994", sig, msg)
	if err != nil {
		t.Errorf("failed to check signature: %d", err)
	}

	if !ok {
		t.Errorf("expected signature to be valid")
	}

	fmt.Println(ok)
}

func TestMiscCrypto(t *testing.T) {
	pk := "4bb214b1f3a0737d758bc3828cdff371e3769fe84a2678da34700cb18d50770e"

	ecdsaPk, err := crypto.ECDSAFromHex(pk)
	if err != nil {
		t.Errorf("error getting ecdsa private key from hex")
	}

	addr := crypto.AddressFromPrivateKey(ecdsaPk)

	if addr != "0x995d95245698212D4Af52c8031F614C3D3127994" {
		t.Errorf("received unexpected address")
	}
}
