package utils

import (
	"encoding/hex"
	"encoding/json"
	"os"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/spf13/cobra"
)

var (
	generateLong    = `Generates a new key ECDSA pair using the secp256k1 curve.`
	generateExample = `kwil-cli utils generate-key`
)

// GenerateKeyCmd returns the command for generating a new key pair.
func generateKeyCmd() *cobra.Command {
	var out string
	var cmd = &cobra.Command{
		Use:     "generate-key",
		Short:   "Generates a new key pair.",
		Long:    generateLong,
		Example: generateExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			pk, err := crypto.GenerateSecp256k1Key()
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			pub := hex.EncodeToString(pk.PubKey().Bytes())
			address, err := auth.EthSecp256k1Authenticator{}.Identifier(pk.PubKey().Bytes())
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if out != "" {
				out, err = common.ExpandPath(out)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				err = os.WriteFile(out, []byte(pk.Hex()), 0644)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				res := privateKeyFileRes{
					PrivateKeyPath: out,
					PublicKey:      pub,
					Address:        address,
				}

				return display.PrintCmd(cmd, &res)
			} else {
				res := privateKeyRes{
					PrivateKey: hex.EncodeToString(pk.Bytes()),
					PublicKey:  pub,
					Address:    address,
				}

				return display.PrintCmd(cmd, &res)
			}
		},
	}

	cmd.Flags().StringVarP(&out, "out", "o", "", "Output file for the generated key pair.")

	return cmd
}

type privateKeyFileRes struct {
	PrivateKeyPath string `json:"private_key_path"`
	PublicKey      string `json:"public_key"`
	Address        string `json:"address"`
}

func (p *privateKeyFileRes) MarshalJSON() ([]byte, error) {
	type res privateKeyFileRes // prevent recursion
	return json.Marshal((*res)(p))
}

func (p *privateKeyFileRes) MarshalText() (text []byte, err error) {
	bts := []byte("Private key written to ")
	bts = append(bts, p.PrivateKeyPath...)
	bts = append(bts, []byte("\nPublic key: ")...)
	bts = append(bts, p.PublicKey...)
	bts = append(bts, []byte("\nAddress: ")...)
	bts = append(bts, p.Address...)
	return bts, nil
}

type privateKeyRes struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}

func (p *privateKeyRes) MarshalJSON() ([]byte, error) {
	type res privateKeyRes // prevent recursion
	return json.Marshal((*res)(p))
}

func (p *privateKeyRes) MarshalText() (text []byte, err error) {
	bts := []byte("Private key: ")
	bts = append(bts, p.PrivateKey...)
	bts = append(bts, []byte("\nPublic key: ")...)
	bts = append(bts, p.PublicKey...)
	bts = append(bts, []byte("\nAddress: ")...)
	bts = append(bts, p.Address...)
	return bts, nil
}
