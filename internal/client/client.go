package client

import (
	"context"
	"fmt"
	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	"github.com/kwilteam/kwil-db/pkg/types"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

//This is used as both a cosmos client and eth client

type CosmosClient struct {
	Client  cosmosclient.Client
	Account *cosmosaccount.Account
	Config  *types.Config
}

type EthClient struct {
}

//goland:noinspection GoUnusedExportedFunction
func NewCosmosClient(ctx context.Context, conf *types.Config) (CosmosClient, error) {
	account := CosmosClient{
		Config: conf,
	}
	cosmos, err := cosmosclient.New(
		ctx,
		cosmosclient.WithAddressPrefix(conf.Wallets.Cosmos.AddressPrefix),
	)
	if err != nil {
		return account, err
	}

	accs, _ := cosmos.AccountRegistry.List()
	fmt.Println(accs)

	accName := "alice"
	acc, err := cosmos.Account(accName)

	if err != nil {
		return account, err
	}
	account.Account = &acc
	account.Client = cosmos
	return account, nil

	// Commenting out below until this keyring shit can get figured out
	/*
		// We need to read in the mnemonic from conf.Wallets.Cosmos.MnemonicPath
		cosmMnemonic, err := os.ReadFile(conf.Wallets.Cosmos.MnemonicPath)
		if err != nil {
			return nil, err
		}

		cosmos, err := cosmosclient.New(
			ctx,
			cosmosclient.WithAddressPrefix(conf.Wallets.Cosmos.AddressPrefix),
		)
		if err != nil {
			return nil, err
		}
		_, err = ImportWallet(&cosmos.AccountRegistry, accName, string(cosmMnemonic))
		if err != nil {
			return nil, err
		}

		acc, err := cosmos.Account(accName)
		if err != nil {
			return nil, err
		}

		return &CosmosClient{
			Account: &acc,
		}, nil
	*/
}

func (c *CosmosClient) Transfer(amt int, toAddr string) error {
	tokenAmt := strconv.Itoa(amt) + "token"
	coins, err := sdk.ParseCoinNormalized(tokenAmt)
	if err != nil {
		return err
	}

	msg := &banktypes.MsgSend{
		FromAddress: c.Account.Address(c.Config.Wallets.Cosmos.AddressPrefix),
		ToAddress:   toAddr,
		Amount:      sdk.Coins{coins},
	}

	txResp, err := c.Client.BroadcastTx(c.Account.Name, msg)
	if err != nil {
		return err
	}

	// Print response from broadcasting a transaction
	fmt.Print("MsgCreatePost:\n\n")
	fmt.Println(txResp)

	return nil
}
