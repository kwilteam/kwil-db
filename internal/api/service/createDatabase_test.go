package service

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/types/chain/pricing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	apitypes "github.com/kwilteam/kwil-db/internal/api/types"
	"github.com/kwilteam/kwil-db/internal/chain/crypto"
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
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				pricing: pb,
			},
			args: args{
				ctx: context.Background(),
				db: &apitypes.CreateDatabaseMsg{
					ID:        string(createDBID(from, name, fee)),
					DBType:    dbt,
					Name:      name,
					Operation: oper,
					Crud:      crd,
					Fee:       fee,
					Signature: tsign(string(createDBID(from, name, fee))),
					From:      from,
				},
			},
			wantErr: false,
		},
		{
			name: "fee too low",
			fields: fields{
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				pricing: pb,
			},
			args: args{
				ctx: context.Background(),
				db: &apitypes.CreateDatabaseMsg{
					ID:        string(createDBID(from, name, "1")),
					DBType:    dbt,
					Name:      name,
					Operation: oper,
					Crud:      crd,
					Fee:       "1",
					Signature: tsign(string(createDBID(from, name, "1"))),
					From:      from,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid signature length",
			fields: fields{
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				pricing: pb,
			},
			args: args{
				ctx: context.Background(),
				db: &apitypes.CreateDatabaseMsg{
					ID:        string(createDBID(from, name, fee)),
					DBType:    dbt,
					Name:      name,
					Operation: oper,
					Crud:      crd,
					Fee:       fee,
					Signature: "ABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCAABCABCABCABCA",
					From:      from,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid signature",
			fields: fields{
				ds:      &MockDepositStore{},
				log:     zerolog.Logger{},
				pricing: pb,
			},
			args: args{
				ctx: context.Background(),
				db: &apitypes.CreateDatabaseMsg{
					ID:        string(createDBID(from, name, fee)),
					DBType:    dbt,
					Name:      name,
					Operation: oper,
					Crud:      crd,
					Fee:       fee,
					Signature: "ABC",
					From:      from,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				ds:      tt.fields.ds,
				log:     tt.fields.log,
				pricing: tt.fields.pricing,
			}
			if err := s.CreateDatabase(tt.args.ctx, tt.args.db); (err != nil) != tt.wantErr {
				t.Errorf("Service.CreateDatabase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

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
