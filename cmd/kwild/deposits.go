package main

import (
	"fmt"
	"kwil/x/cfgx"
	chainClient "kwil/x/chain/client"
	"kwil/x/crypto"
	app "kwil/x/deposits/app"
	"kwil/x/deposits/dto"
	deposits "kwil/x/deposits/service"
	"kwil/x/sqlx/sqlclient"
	"os"
)

func buildDeposits(cfg cfgx.Config, db *sqlclient.DB, cc chainClient.ChainClient, privateKey string) (*app.Service, error) {
	config := cfg.Select("deposits")
	escrowAddr := config.GetString("escrow-address", "")
	if escrowAddr == "" {
		return nil, fmt.Errorf("escrow-address must be set.  received empty string")
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

	svc, err := deposits.NewService(dto.Config{
		EscrowAddress: escrowAddr,
	}, db, cc, pk)
	if err != nil {
		return nil, fmt.Errorf("error creating deposit service: %d", err)
	}

	return app.NewService(svc), nil
}
