package config_test

import (
	"fmt"
	config "kwil/internal/app/kwild/config"
	"kwil/pkg/crypto"
	"os"
	"testing"
)

func Test_Config(t *testing.T) {
	os.Setenv("KWILD_PRIVATE_KEY", "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e")
	os.Setenv("KWILD_PORT", "8081")
	os.Setenv("KWILD_DEPOSITS_POOL_ADDRESS", "0xabc")
	cfg, err := config.LoadKwildConfig()
	if err != nil {
		t.Fatal(err)
	}

	addr := crypto.AddressFromPrivateKey(cfg.PrivateKey)
	fmt.Println(addr)
}
