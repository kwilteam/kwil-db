package service

import (
	"math/big"
	"reflect"
	"testing"

	kconf "github.com/kwilteam/kwil-db/internal/config/test"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Commenting out for now since the constructor keeps changing
/*func TestNewService(t *testing.T) {
	type args struct {
		conf *types.Config
		ds   DepositStore
	}
	tests := []struct {
		name string
		args args
		want *Service
	}{
		{
			name: "valid use",
			args: args{
				conf: &types.Config{},
				ds:   &MockDepositStore{},
			},
			want: &Service{
				conf: &types.Config{},
				ds:   &MockDepositStore{},
				log:  log.With().Str("module", "service").Logger(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewService(tt.args.conf, tt.args.ds); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewService() = %v, want %v", got, tt.want)
			}
		})
	}
}*/

func TestService_validateBalances(t *testing.T) {
	fr := "0x995d95245698212D4Af52c8031F614C3D3127994"
	f := "5"
	lf := "1"
	conf := kconf.GetTestConfig(t)

	type fields struct {
		conf *types.Config
		ds   DepositStore
		log  zerolog.Logger
	}
	type args struct {
		from *string
		op   *string
		f    *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *big.Int
		wantErr bool
	}{
		{
			name: "valid amount",
			fields: fields{
				conf: conf,
				ds:   &MockDepositStore{},
				log:  log.With().Str("module", "service").Logger(),
			},
			args: args{
				from: &fr,
				op:   &conf.Cost.Database.Create,
				f:    &f,
			},
			want:    big.NewInt(0),
			wantErr: false,
		},
		{
			name: "invalid amount",
			fields: fields{
				conf: conf,
				ds:   &MockDepositStore{},
				log:  log.With().Str("module", "service").Logger(),
			},
			args: args{
				from: &fr,
				op:   &conf.Cost.Database.Create,
				f:    &lf,
			},
			want:    big.NewInt(0),
			wantErr: true,
		},
		{
			name: "valid amount (different operation)",
			fields: fields{
				conf: conf,
				ds:   &MockDepositStore{},
				log:  log.With().Str("module", "service").Logger(),
			},
			args: args{
				from: &fr,
				op:   &conf.Cost.Database.Delete,
				f:    &f,
			},
			want:    big.NewInt(2),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				conf: tt.fields.conf,
				ds:   tt.fields.ds,
				log:  tt.fields.log,
			}
			got, err := s.validateBalances(tt.args.from, tt.args.op, tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.validateBalances() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.String(), tt.want.String()) {
				t.Errorf("Service.validateBalances() got1 = %v, want %v", got, tt.want)
			}
		})
	}
}
