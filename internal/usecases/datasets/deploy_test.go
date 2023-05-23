package datasets_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/internal/entity/data"
	"github.com/kwilteam/kwil-db/internal/usecases/datasets"
	"github.com/kwilteam/kwil-db/pkg/tx"
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
			u, err := newDatasetUsecase(tt.args.ctx)
			if err != nil {
				t.Error(err)
			}

			got, err := u.Deploy(tt.args.ctx, tt.args.deployment)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatasetUseCase.Deploy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatasetUseCase.Deploy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newDatasetUsecase(ctx context.Context) (datasets.DatasetUseCaseInterface, error) {
	return datasets.New(ctx,
		datasets.WithEngine(newMockEngine()),
		datasets.WithAccountStore(&mockAccountStore{}),
	)
}
