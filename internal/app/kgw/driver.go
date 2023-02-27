package kgw

import (
	"context"
	"fmt"
	"kwil/internal/app/kwild"
	"kwil/internal/pkg/graphql/query"
)

// KgwDriver is a driver for the gw client for integration tests
type KgwDriver struct {
	kwild.KwildDriver

	gatewayAddr string // to ignore the gatewayAddr returned by the config.service
}

func NewKgwDriver(gatewayAddr string, kwildDriver *kwild.KwildDriver) *KgwDriver {
	return &KgwDriver{
		KwildDriver: *kwildDriver,
		gatewayAddr: gatewayAddr,
	}
}

func (d *KgwDriver) QueryDatabase(ctx context.Context, queryStr string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/graphql", d.gatewayAddr)
	return query.Query(ctx, url, queryStr)
}
