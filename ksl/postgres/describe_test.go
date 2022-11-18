package postgres_test

import (
	"context"
	_ "ksl/postgres"
	"ksl/sqlclient"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDescribe(t *testing.T) {
	driver, err := sqlclient.Open("postgres://localhost:5432/postgres?sslmode=disable&schema=public")
	require.NoError(t, err)

	sch, err := driver.DescribeContext(context.Background(), "public")
	require.NoError(t, err)

	_ = sch
}
