package client

import (
	"context"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	"github.com/kwilteam/kwil-db/pkg/types"
	"os"
)

//var ErrAccountExists = errors.New("account already exists")

// ImportWallet takes a mnemonic and adds a new account to the keyring.  It also returns the account.
func importWallet(ctx context.Context, conf *types.Config) (cosmosclient.Client, error) {

	cosmos, err := cosmosclient.New(
		ctx,
		cosmosclient.WithAddressPrefix(conf.Wallets.Cosmos.AddressPrefix), // TODO: should make an in-memory backend and then add address later
	)

	if err != nil {
		return cosmos, err
	}

	mn, err := os.ReadFile(conf.Wallets.Cosmos.MnemonicPath)
	if err != nil {
		return cosmos, err
	}

	cosmos.AccountRegistry.Import(conf.Wallets.Cosmos.KeyName, string(mn), "")
	cosmos.Factory.Prepare(cosmos.Context()) // This is fucking stupid.

	return cosmos, nil
}

/*func getSignAlgo(r cosmosaccount.Registry) (keyring.SignatureAlgo, error) {
	algos, _ := r.Keyring.SupportedAlgorithms()
	return keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), algos)
}

func hdPath() string {
	return hd.CreateHDPath(sdktypes.GetConfig().GetCoinType(), 0, 0).String()
}
*/
