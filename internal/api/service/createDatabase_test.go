package service

import (
	"context"
<<<<<<< HEAD
	"encoding/json"
	"github.com/kwilteam/kwil-db/pkg/pricing"
	"math/big"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	apitypes "github.com/kwilteam/kwil-db/internal/api/types"
	"github.com/kwilteam/kwil-db/internal/crypto"
	"github.com/rs/zerolog"
)

type MockCruds struct {
	Create string
	Modify string
	Delete string
}

type MockPricingStruct struct {
	Database MockCruds
	Table    MockCruds
	Role     MockCruds
	Query    MockCruds
}

func getTestPriceBuilder() pricing.PriceBuilder {
	pb := MockPricingStruct{
		Database: MockCruds{
			Create: "3",
			Modify: "-1",
			Delete: "-1",
		},
		Table: MockCruds{
			Create: "2",
			Modify: "3",
			Delete: "1",
		},
		Role: MockCruds{
			Create: "2",
			Modify: "-1",
			Delete: "1",
		},
		Query: MockCruds{
			Create: "2",
			Modify: "3",
			Delete: "-1",
		},
	}

	// convert pb to bytes
	b, err := json.Marshal(pb)
	if err != nil {
		panic(err)
	}

	p, err := pricing.New(b)
	if err != nil {
		panic(err)
	}
	return p
}

// function to allow on-the-fly signing (without error handling)
func tsign(val string) string {
	pk, err := ethcrypto.HexToECDSA("4bb214b1f3a0737d758bc3828cdff371e3769fe84a2678da34700cb18d50770e")
	if err != nil {
		panic(err)
	}

	sig, err := crypto.Sign([]byte(val), pk)
	if err != nil {
		panic(err)
	}

	return sig
}

// these are standard variables for testing
var from = "0x995d95245698212D4Af52c8031F614C3D3127994"
var name = "testdb"
var fee = "5"
var dbt = "postgres"
var oper byte = 0
var crd byte = 0

func TestService_CreateDatabase(t *testing.T) {
	pb := getTestPriceBuilder()

	type fields struct {
		ds      DepositStore
		log     zerolog.Logger
		pricing pricing.PriceBuilder
	}
	type args struct {
		ctx context.Context
		db  *apitypes.CreateDatabaseMsg
=======
	"math/big"
	"testing"

	kconf "github.com/kwilteam/kwil-db/internal/config/test"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog"
)

type MockDepositStore struct {
	bal *big.Int
}

func (m *MockDepositStore) GetBalance(address string) (*big.Int, error) {
	// if the big.Int is nil, set to 5
	if m.bal == nil {
		m.bal = big.NewInt(5)
	}

	return m.bal, nil
}

func (m *MockDepositStore) SetBalance(address string, balance *big.Int) error {
	m.bal = balance
	return nil
}

type MockCosmosClient struct {
}

func (m *MockCosmosClient) CreateDB(db *types.CreateDatabase) error {
	return nil
}

func TestService_CreateDatabase(t *testing.T) {
	type fields struct {
		conf    *types.Config
		ds      DepositStore
		log     zerolog.Logger
		cClient CosmosClient
	}
	type args struct {
		ctx context.Context
		db  *types.CreateDatabase
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid request",
			fields: fields{
<<<<<<< HEAD
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				pricing: pb,
			},
			args: args{
				ctx: context.Background(),
				db: &apitypes.CreateDatabaseMsg{
					ID:        string(createDBID(from, name, fee, dbt)),
					DBType:    dbt,
					Name:      name,
					Operation: oper,
					Crud:      crd,
					Fee:       fee,
					Signature: tsign(string(createDBID(from, name, fee, dbt))),
					From:      from,
=======
				conf:    kconf.GetTestConfig(t),
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				cClient: &MockCosmosClient{},
			},
			args: args{
				ctx: context.Background(),
				db: &types.CreateDatabase{
					Id:        "kwil",
					DBType:    "test",
					Name:      "testdb",
					Fee:       "5",
					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
				},
			},
			wantErr: false,
		},
		{
			name: "fee too low",
			fields: fields{
<<<<<<< HEAD
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				pricing: pb,
			},
			args: args{
				ctx: context.Background(),
				db: &apitypes.CreateDatabaseMsg{
					ID:        string(createDBID(from, name, "1", dbt)),
					DBType:    dbt,
					Name:      name,
					Operation: oper,
					Crud:      crd,
					Fee:       "1",
					Signature: tsign(string(createDBID(from, name, "1", dbt))),
					From:      from,
=======
				conf:    kconf.GetTestConfig(t),
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				cClient: &MockCosmosClient{},
			},
			args: args{
				ctx: context.Background(),
				db: &types.CreateDatabase{
					Id:        "kwil",
					DBType:    "test",
					Name:      "testdb",
					Fee:       "1",
					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
				},
			},
			wantErr: true,
		},
		{
			name: "invalid signature length",
			fields: fields{
<<<<<<< HEAD
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				pricing: pb,
			},
			args: args{
				ctx: context.Background(),
				db: &apitypes.CreateDatabaseMsg{
					ID:        string(createDBID(from, name, fee, dbt)),
					DBType:    dbt,
					Name:      name,
					Operation: oper,
					Crud:      crd,
					Fee:       fee,
					Signature: "ABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCA",
					From:      from,
=======
				conf:    kconf.GetTestConfig(t),
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				cClient: &MockCosmosClient{},
			},
			args: args{
				ctx: context.Background(),
				db: &types.CreateDatabase{
					Id:        "kwil",
					DBType:    "test",
					Name:      "testdb",
					Fee:       "5",
					Signature: "0x39fd0a55rr51cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
				},
			},
			wantErr: true,
		},
		{
			name: "invalid signature",
			fields: fields{
<<<<<<< HEAD
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				pricing: pb,
			},
			args: args{
				ctx: context.Background(),
				db: &apitypes.CreateDatabaseMsg{
					ID:        string(createDBID(from, name, fee, dbt)),
					DBType:    dbt,
					Name:      name,
					Operation: oper,
					Crud:      crd,
					Fee:       fee,
					Signature: "ABC",
					From:      from,
=======
				conf:    kconf.GetTestConfig(t),
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				cClient: &MockCosmosClient{},
			},
			args: args{
				ctx: context.Background(),
				db: &types.CreateDatabase{
					Id:        "kwilll",
					DBType:    "test",
					Name:      "testdb",
					Fee:       "5",
					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
<<<<<<< HEAD
				ds:      tt.fields.ds,
				log:     tt.fields.log,
				pricing: tt.fields.pricing,
=======
				conf:    tt.fields.conf,
				ds:      tt.fields.ds,
				log:     tt.fields.log,
				cClient: tt.fields.cClient,
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
			}
			if err := s.CreateDatabase(tt.args.ctx, tt.args.db); (err != nil) != tt.wantErr {
				t.Errorf("Service.CreateDatabase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
<<<<<<< HEAD

type MockDepositStore struct {
	bal *big.Int
}

func (m *MockDepositStore) GetBalance(address string) (*big.Int, error) {
	// if the big.Int is nil, set to 5
	if m.bal == nil {
		m.bal = big.NewInt(5)
	}

	return m.bal, nil
}

func (m *MockDepositStore) SetBalance(address string, balance *big.Int) error {
	m.bal = balance
	return nil
}
=======
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
