package specifications

import (
	"context"
	"kwil/x/types/databases"
	"kwil/x/types/transactions"
	"testing"

	anytype "kwil/x/types/data_types/any_type"
)

type QDeployDatabase interface {
	DeployDatabase(ctx context.Context, db *databases.Database[anytype.KwilAny]) (*transactions.Response, error)
}

func DeploySpecification(t *testing.T, deploy QDeployDatabase, db *databases.Database[anytype.KwilAny], wantErr bool, want *transactions.Response) {
	res, err := deploy.DeployDatabase(context.Background(), db)
	if err != nil && !wantErr {
		t.Fatal(err)
	}

	if wantErr && err == nil {
		t.Fatal("expected error, got nil")
	}

	if want != nil && res != want {
		t.Fatalf("expected %v, got %v", want, res)
	}

	t.Log(res.Hash)
}
