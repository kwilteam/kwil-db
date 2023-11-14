package server

import (
	"testing"

	bClient "github.com/kwilteam/kwil-db/core/bridge/client"
	"github.com/kwilteam/kwil-db/core/bridge/syncer"
	"github.com/kwilteam/kwil-db/core/types/chain"
	"github.com/kwilteam/kwil-db/internal/chainsyncer"
)

func Test_BuildChainSyncer(t *testing.T) {
	escrowAddr := "0xcc46cc0960d6903a5b7a76d431aed56fef70e7b0"
	tokenAddr := "0x8ce9d23b427b80ab5e21c272a46acd3a27082836"
	bridgeClient, err := bClient.New("http://localhost:8545", chain.GOERLI, tokenAddr, escrowAddr)
	if err != nil {
		failBuild(err, "failed to build bridge client")
	}

	// build block syncer
	blockSyncer, err := syncer.New(bridgeClient.ChainClient())
	if err != nil {
		failBuild(err, "failed to build block syncer")
	}

	// build chain syncer
	chainSyncer := chainsyncer.New(bridgeClient, blockSyncer, nil)
	chainSyncer.Start()
}
