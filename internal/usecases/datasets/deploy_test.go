package datasets_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/internal/entity/data"
	"github.com/kwilteam/kwil-db/internal/usecases/datasets"
	accountTesting "github.com/kwilteam/kwil-db/pkg/balances/testing"
	"github.com/kwilteam/kwil-db/pkg/engine"
	engineTesting "github.com/kwilteam/kwil-db/pkg/engine/testing"
	"github.com/kwilteam/kwil-db/pkg/tx"
	"github.com/stretchr/testify/assert"
)

func TestDatasetUseCase_Deploy(t *testing.T) {
	type args struct {
		ctx        context.Context
		deployment *entity.DeployDatabase
	}
	tests := []struct {
		name    string
		args    args
		want    *tx.Receipt
		wantErr bool
	}{
		{
			name: "successful deployment",
			args: args{
				ctx: context.Background(),
				deployment: &entity.DeployDatabase{
					Schema: &entity.Schema{
						Owner: "0x123",
						Name:  "test",
						Tables: []*entity.Table{
							data.TableUsers,
						},
						Actions: []*entity.Action{
							data.ActionInsertUser,
						},
					},
					Tx: &tx.Transaction{
						Hash:   []byte("0xabc"),
						Sender: "0x123",
						Nonce:  0,
						Fee:    "1000000000000000000",
					},
				},
			},
			want: &tx.Receipt{
				TxHash: []byte("0xabc"),
				Fee:    "1000000000000000000",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			eng, teardown, err := engineTesting.NewTestEngine(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer teardown()

			accountStore, td, err := accountTesting.NewTestAccountStore(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer td()

			u, err := datasets.New(ctx,
				datasets.WithEngine(eng),
				datasets.WithAccountStore(accountStore),
			)
			if err != nil {
				t.Fatal(err)
			}

			got, err := u.Deploy(tt.args.ctx, tt.args.deployment)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatasetUseCase.Deploy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatasetUseCase.Deploy() = %v, want %v", got, tt.want)
			}

			schema, err := u.GetSchema(ctx, engine.GenerateDBID(tt.args.deployment.Schema.Name, tt.args.deployment.Schema.Owner))
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.args.deployment.Schema, schema)
		})
	}
}
