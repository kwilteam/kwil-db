package service

// import (
// 	"context"
// 	"testing"

// 	kconf "github.com/kwilteam/kwil-db/internal/config/test"
// 	types "github.com/kwilteam/kwil-db/pkg/types/chain"
// 	"github.com/rs/zerolog"
// )

// func TestService_DDL(t *testing.T) {
// 	conf := kconf.GetTestConfig(t)
// 	type fields struct {
// 		conf *types.Config
// 		ds   DepositStore
// 		log  zerolog.Logger
// 	}
// 	type args struct {
// 		ctx context.Context
// 		ddl *types.DDL
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "valid table create request",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwil",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "table_create",
// 					DDL:       "create table test_table (column1 int)",
// 					Fee:       "3",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "valid table delete request",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwil",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "table_delete",
// 					DDL:       "create table test_table (column1 int)",
// 					Fee:       "3",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "valid table modify request",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwil",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "table_modify",
// 					DDL:       "create table test_table (column1 int)",
// 					Fee:       "3",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "valid query create request",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwil",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "query_create",
// 					DDL:       "create query...",
// 					Fee:       "2",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "valid query delete request",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwil",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "query_delete",
// 					DDL:       "delete query...",
// 					Fee:       "1",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "valid role create request",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwil",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "role_create",
// 					DDL:       "delete role...",
// 					Fee:       "2",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "valid role modify request",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwil",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "role_modify",
// 					DDL:       "delete role...",
// 					Fee:       "1",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "valid role delete request",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwil",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "role_delete",
// 					DDL:       "delete role ...",
// 					Fee:       "1",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "invalid ddl type",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwil",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "table_update",
// 					DDL:       "create table test_table (column1 int)",
// 					Fee:       "3",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "missing from address",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwil",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "table_create",
// 					DDL:       "create table test_table (column1 int)",
// 					Fee:       "3",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "invalid signature",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwiller",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "table_create",
// 					DDL:       "create table test_table (column1 int)",
// 					Fee:       "3",
// 					Signature: "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
// 					From:      "0x995d95245698212D4Af52c8031F614C3D3127994",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "fee too low",
// 			fields: fields{
// 				conf: conf,
// 				ds:   &MockDepositStore{},
// 				log:  zerolog.Logger{},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ddl: &types.DDL{
// 					Id:        "kwil",
// 					Name:      "testdb",
// 					Owner:     "kwil",
// 					DBType:    "postgres",
// 					Type:      "table_create",
// 					DDL:       "create table test_table (column1 int)",
// 					Fee:       "1",
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
// 				conf: tt.fields.conf,
// 				ds:   tt.fields.ds,
// 				log:  tt.fields.log,
// 			}
// 			if err := s.DDL(tt.args.ctx, tt.args.ddl); (err != nil) != tt.wantErr {
// 				t.Errorf("Service.DDL() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
