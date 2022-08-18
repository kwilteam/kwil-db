package client

import (
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
)

var ErrAccountExists = errors.New("account already exists")

// ImportWallet takes a mnemonic and adds a new account to the keyring.  It also returns the account.
func ImportWallet(r *cosmosaccount.Registry, name, mnemonic string) (cosmosaccount.Account, error) {

	// Check if account exists
	fmt.Println(1)
	acc, err := r.GetByName(name)
	if err == nil { // if so, delete it
		r.DeleteByName(name)
		r.Keyring.Delete(name)
	}

	r.Keyring.Delete(name)
	fmt.Println(2)
	algo, err := getSignAlgo(*r)
	if err != nil {
		return acc, err
	}

	fmt.Println(3)
	info, err := r.Keyring.NewAccount(name, mnemonic, "", hdPath(), algo)
	if err != nil {
		return acc, err
	}

	fmt.Println(4)
	acc = cosmosaccount.Account{
		Name: name,
		Info: info,
	}

	fmt.Println(5)

	return acc, nil
}

func getSignAlgo(r cosmosaccount.Registry) (keyring.SignatureAlgo, error) {
	algos, _ := r.Keyring.SupportedAlgorithms()
	return keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), algos)
}

func hdPath() string {
	return hd.CreateHDPath(sdktypes.GetConfig().GetCoinType(), 0, 0).String()
}
