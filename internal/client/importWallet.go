package client

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
)

var ErrAccountExists = errors.New("account already exists")

// ImportWallet takes a mnemonic and adds a new account to the keyring.  It also returns the account.
func ImportWallet(r *cosmosaccount.Registry, name, mnemonic string) (cosmosaccount.Account, error) {

	// Check if account exists
	acc, err := r.GetByName(name)
	if err == nil { // if so, delete it
		_ = r.DeleteByName(name)
		_ = r.Keyring.Delete(name)
	}

	_ = r.Keyring.Delete(name)
	algo, err := getSignAlgo(*r)
	if err != nil {
		return acc, err
	}

	info, err := r.Keyring.NewAccount(name, mnemonic, "", hdPath(), algo)
	if err != nil {
		return acc, err
	}

	acc = cosmosaccount.Account{
		Name: name,
		Info: info,
	}

	return acc, nil
}

func getSignAlgo(r cosmosaccount.Registry) (keyring.SignatureAlgo, error) {
	algos, _ := r.Keyring.SupportedAlgorithms()
	return keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), algos)
}

func hdPath() string {
	return hd.CreateHDPath(sdktypes.GetConfig().GetCoinType(), 0, 0).String()
}
