package database

import (
	"context"
	"encoding/json"
	"fmt"
	"kwil/x/cli/chain"
	"kwil/x/cli/util"
	"kwil/x/crypto"
	"kwil/x/proto/apipb"
	"kwil/x/sqlx/models"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func dropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Drops a database",
		Long:  "Drops a database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client := apipb.NewKwilServiceClient(cc)
				// should be one arg
				if len(args) != 1 {
					return fmt.Errorf("deploy requires one argument: database name")
				}

				c, err := chain.NewClientV(viper.GetViper())
				if err != nil {
					return err
				}

				// if this all works, lets construct the request
				nonce := util.GenerateNonce(32)
				fee := "1000000000000000000" // TODO: should call the estimate endpoint

				// construct payload
				payload := models.DropDatabase{
					Name:  args[0],
					Owner: c.Address.String(),
				}

				// marshal payload
				bts, err := json.Marshal(payload)
				if err != nil {
					return err
				}

				// add message type
				bts = append([]byte{models.DROP_DATABASE.Byte()}, bts...)

				// add version
				bts = append([]byte{0}, bts...)

				// construct ID
				// first hash the payload
				payloadHash := crypto.Sha384Str(bts)

				sb := strings.Builder{}
				sb.WriteString(payloadHash)
				sb.WriteString(fee)
				sb.WriteString(nonce)
				sb.WriteString(c.Address.String())
				id := crypto.Sha384Str([]byte(sb.String()))

				// sign it
				sig, err := crypto.Sign([]byte(id), c.PrivateKey)
				if err != nil {
					return err
				}

				// construct the request
				req := &apipb.DropSchemaRequest{
					Tx: &apipb.Tx{
						Id:        id,
						Payload:   bts,
						Fee:       fee,
						Nonce:     nonce,
						Signature: sig,
						Sender:    c.Address.String(),
					},
				}

				// send it
				resp, err := client.DropSchema(ctx, req)

				if err != nil {
					return err
				}

				fmt.Println(resp)

				return nil
			})
		},
	}
	return cmd
}
