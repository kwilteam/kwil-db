package config_test

import (
	"os"
	"testing"

	config "github.com/kwilteam/kwil-db/internal/app/kwild/config"
)

func Test_Config(t *testing.T) {
	os.Setenv("KWILD_PRIVATE_KEY", "f2d82d73ba03a7e843443f2b3179a01398144baa4a23d40d1e8a3a8e4fb217d0484d59f4de46b2174ebce66ac3afa7989b444244323c19a74b683f54cf33227c")
	os.Setenv("KWILD_PORT", "8081")
	os.Setenv("KWILD_EXTENSION_ENDPOINTS", "localhost:8080,localhost:8081,    localhost:8082")
	_, err := config.LoadKwildConfig()
	if err != nil {
		t.Fatal(err)
	}
}
