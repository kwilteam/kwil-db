package client

import (
	kwiltypes "github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"
	"github.com/kwilteam/kwil-db/pkg/types"
	"strings"
)

func (c *CosmosClient) CreateDB(db *types.CreateDatabase) error {
	sb := strings.Builder{}
	sb.WriteString(db.From)
	sb.WriteString("/")
	sb.WriteString(db.Name)

	msg := kwiltypes.NewMsgCreateDatabase(c.address, sb.String())

	txResp, err := c.Client.BroadcastTx(c.conf.Wallets.Cosmos.KeyName, msg)
	if err != nil {
		return err
	}

	c.log.Info().Msgf("CreateDB response from cosmos: %+v", txResp)

	return nil
}
