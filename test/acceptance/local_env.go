package acceptance

import (
	"fmt"
	"os"
)

var envTemplage = `
KWILD_URL=%s
KGW_URL=%s
TEST_USER_PK=%s
TEST_USER_ADDR=%s
TEST_DEPLOYER_PK=%s
TEST_DEPLOYER_ADDR=%s
`

func DumpEnv(cfg *TestEnvCfg) {
	content := fmt.Sprintf(envTemplage,
		cfg.NodeURL, cfg.GatewayURL,
		cfg.UserPrivateKeyString, cfg.UserAddr,
		cfg.DatabaseDeployerPrivateKeyString, cfg.DeployerAddr,
	)

	err := os.WriteFile("../../.local_env", []byte(content), 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to write file: %v", err))
	}
}
