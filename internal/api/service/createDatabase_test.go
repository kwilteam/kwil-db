package service

// import (
// 	"context"
// 	"math/big"
// 	"testing"

// 	kconf "github.com/kwilteam/kwil-db/internal/config/test"
// 	types "github.com/kwilteam/kwil-db/pkg/types/chain"
// 	"github.com/rs/zerolog"
// )

// type MockDepositStore struct {
// 	bal *big.Int
// }

// func (m *MockDepositStore) GetBalance(address string) (*big.Int, error) {
// 	// if the big.Int is nil, set to 5
// 	if m.bal == nil {
// 		m.bal = big.NewInt(5)
// 	}

// 	return m.bal, nil
// }

// func (m *MockDepositStore) SetBalance(address string, balance *big.Int) error {
// 	m.bal = balance
// 	return nil
// }

// type MockCosmosClient struct {
// }

// func (m *MockCosmosClient) CreateDB(db *types.CreateDatabase) error {
// 	return nil
// }

// func TestService_CreateDatabase(t *testing.T) {
// 	type fields struct {
// 		conf    *types.Config
// 		ds      DepositStore
// 		log     zerolog.Logger
// 		cClient CosmosClient
// 	}
// 	type args struct {
// 		ctx context.Context
// 		db  *types.CreateDatabase
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "valid request",
// 			fields: fields{
// 				conf:    kconf.GetTestConfig(t),
// 				ds:      &MockDepositStore{},
// 				log:     zerolog.Logger{},
// 				cClient: &MockCosmosClient{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				db: &types.CreateDatabase{
// 					Id:        "kwil",
// 					DBType:    "test",
// 					Name:      "testdb",
// 					Fee:       "5",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "fee too low",
// 			fields: fields{
// 				conf:    kconf.GetTestConfig(t),
// 				ds:      &MockDepositStore{},
// 				log:     zerolog.Logger{},
// 				cClient: &MockCosmosClient{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				db: &types.CreateDatabase{
// 					Id:        "kwil",
// 					DBType:    "test",
// 					Name:      "testdb",
// 					Fee:       "1",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "invalid signature length",
// 			fields: fields{
// 				conf:    kconf.GetTestConfig(t),
// 				ds:      &MockDepositStore{},
// 				log:     zerolog.Logger{},
// 				cClient: &MockCosmosClient{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				db: &types.CreateDatabase{
// 					Id:        "kwil",
// 					DBType:    "test",
// 					Name:      "testdb",
// 					Fee:       "5",
// 					Signature: "0x39fd0a55rr51cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "invalid signature",
// 			fields: fields{
// 				conf:    kconf.GetTestConfig(t),
// 				ds:      &MockDepositStore{},
// 				log:     zerolog.Logger{},
// 				cClient: &MockCosmosClient{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				db: &types.CreateDatabase{
// 					Id:        "kwilll",
// 					DBType:    "test",
// 					Name:      "testdb",
// 					Fee:       "5",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := &Service{
// 				conf:    tt.fields.conf,
// 				ds:      tt.fields.ds,
// 				log:     tt.fields.log,
// 				cClient: tt.fields.cClient,
// 			}
// 			if err := s.CreateDatabase(tt.args.ctx, tt.args.db); (err != nil) != tt.wantErr {
// 				t.Errorf("Service.CreateDatabase() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
