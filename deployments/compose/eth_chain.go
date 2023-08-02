package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/kwilteam/kwil-db/test/acceptance"
	"github.com/kwilteam/kwil-db/test/acceptance/utils/deployer"
)

var (
	chainRpc    = "http://127.0.0.1:8545"
	deployerPK  = "dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5"
	userPK      = "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"
	userAccount = "0xc89D42189f0450C2b2c3c61f58Ec5d628176A1E7"
)

func getChainDeployer() deployer.Deployer {
	chainDeployer := acceptance.GetDeployer("eth", chainRpc, deployerPK, big.NewInt(10))
	return chainDeployer
}

func initContract(chainDeployer deployer.Deployer) {
	// deploy token and escrow contract
	ctx := context.Background()
	tokenAddress, err := chainDeployer.DeployToken(ctx)
	if err != nil {
		log.Fatal(err)
	}
	escrowAddress, err := chainDeployer.DeployEscrow(ctx, tokenAddress.String())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("token address:", tokenAddress.String())
	fmt.Println("pool address:", escrowAddress.String())

	if err := chainDeployer.FundAccount(ctx, userAccount, 9000000000000000000); err != nil {
		log.Fatal(err)
	}
}

func keepMiningBlocks(chainDeployer deployer.Deployer) {
	ctx := context.Background()
	for {
		time.Sleep(3 * time.Second)
		// to mine new blocks
		err := chainDeployer.FundAccount(ctx, userAccount, 1)
		if err != nil {
			fmt.Println("funded user account failed", err)
		}
	}
}

func main() {
	fmt.Println("eth_chain PID:", os.Getpid())
	chainDeployer := getChainDeployer()
	initContract(chainDeployer)
	keepMiningBlocks(chainDeployer)
}
