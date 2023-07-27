package arweave

import (
	"crypto/ecdsa"
	"encoding/json"
	"time"

	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/goether"
	"github.com/kwilteam/kwil-db/pkg/crypto"
)

const ticker = "matic"

type BundlrClient struct {
	bundlrEndpoint string
	signer         *goar.ItemSigner
	tags           []types.Tag
}

func NewBundlrClient(bundlrEndpoint string, privateKey *ecdsa.PrivateKey) (*BundlrClient, error) {
	signer, err := goether.NewSigner(crypto.HexFromECDSAPrivateKey(privateKey)) // ecdsa signer
	if err != nil {
		return nil, err
	}

	itemSigner, err := goar.NewItemSigner(signer)
	if err != nil {
		return nil, err
	}

	return &BundlrClient{
		bundlrEndpoint: bundlrEndpoint,
		signer:         itemSigner,
	}, nil
}

func (c *BundlrClient) StoreItem(bts []byte) (*BundlrResponse, error) {
	uniqueItem := &dataItem{
		Data:      bts,
		Timestamp: time.Now().UnixNano(),
	}

	dataBts, err := json.Marshal(uniqueItem)
	if err != nil {
		return nil, err
	}

	item, err := c.signer.CreateAndSignItem(dataBts, "", "", c.tags)
	if err != nil {
		return nil, err
	}

	resp, err := submitItemToBundlr(item, c.bundlrEndpoint, ticker)
	if err != nil {
		return nil, err
	}

	return &BundlrResponse{
		TxID: resp.Id,
	}, nil
}

type dataItem struct {
	Data      []byte `json:"data"`
	Timestamp int64  `json:"timestamp"` // local node timestamp
}

type BundlrResponse struct {
	TxID string `json:"tx_id"`
}

type Tag struct {
	Name  string
	Value string
}
