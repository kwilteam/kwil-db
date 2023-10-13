package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/internal/abci"

	"github.com/alexflint/go-arg"
)

// kwil-admin key info
// kwil-admin key gen

type KeyCmd struct {
	Info *KeyInfoCmd `arg:"subcommand:info" help:"Display info about a node private key."`
	Gen  *KeyGenCmd  `arg:"subcommand:gen" help:"Generate a node private key."`
}

type KeyInfoCmd struct {
	PrivKey HexArg `arg:"positional" help:"Private key (hexadecimal string)"`

	PrivKeyFile string `arg:"-k,--key-file" help:"file containing the private key"`
}

type KeyGenCmd struct {
	PrivKeyFile string `arg:"-o,--key-file" help:"file to which the new private key is written (stdout by default)"`
	Raw         bool   `arg:"-R,--raw" help:"just print the private key hex without other encodings, public key, or node ID"`
}

func keyFromBytesOrFile(key []byte, keyFile string) ([]byte, error) {
	if len(key) > 0 {
		return key, nil
	}
	if keyFile == "" {
		return nil, errors.New("must provide with the private key file or hex string")
	}
	return abci.ReadKeyFile(keyFile)
}

func (kc *KeyCmd) run(ctx context.Context) error {
	switch {
	case kc.Info != nil:
		privKey, err := keyFromBytesOrFile(kc.Info.PrivKey, kc.Info.PrivKeyFile)
		if err != nil {
			return err
		}
		abci.PrintPrivKeyInfo(privKey)
		return nil

	case kc.Gen != nil:
		privKey := abci.GeneratePrivateKey()
		if kc.Gen.PrivKeyFile == "" {
			if kc.Gen.Raw {
				fmt.Println(hex.EncodeToString(privKey))
			} else {
				abci.PrintPrivKeyInfo(privKey)
			}
			return nil
		}
		keyHex := hex.EncodeToString(privKey[:])
		return os.WriteFile(kc.Gen.PrivKeyFile, []byte(keyHex), 0600)

	default:
		return arg.ErrHelp
	}
}
