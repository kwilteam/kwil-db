package server

import (
	"fmt"
	"kwil/kwil/repository"
	chainClient "kwil/pkg/chain/client"
	"kwil/pkg/sql/sqlclient"
	"kwil/x/cfgx"
	"kwil/x/crypto"
	"kwil/x/deposits"
	depositTypes "kwil/x/types/deposits"
	"os"
)

func buildDeposits(cfg cfgx.Config, db *sqlclient.DB, queries repository.Queries, cc chainClient.ChainClient) (deposits.Depositer, error) {
	config := cfg.Select("deposits")
	escrowAddr := config.GetString("escrow-address", "")
	if escrowAddr == "" {
		return nil, fmt.Errorf("escrow-address must be set.  received empty string")
	}

	depositCfg := cfg.Select("deposit-settings")
	keysCfg := depositCfg.Select("keys")
	fmt.Printf("keysCfg: %+v\n", keysCfg.ToMap())
	privateKey := keysCfg.GetString("wallet-key", "")
	if privateKey == "" {
		return nil, fmt.Errorf("wallet-key must be set.  received empty string")
	}

	chunkSize, err := config.GetInt64("chunk-size", 100000)
	if err != nil {
		return nil, fmt.Errorf("error getting chunk-size from config: %d", err)
	}

	os.Setenv("deposit_chunk_size", fmt.Sprint(chunkSize))

	// convert private key to ecdsa
	pk, err := crypto.ECDSAFromHex(privateKey)
	if err != nil {
		return nil, fmt.Errorf("error converting private key to ecdsa: %d", err)
	}

	return deposits.NewDepositer(depositTypes.Config{
		EscrowAddress: escrowAddr,
	}, db, queries, cc, pk)
}
