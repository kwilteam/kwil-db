package config

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/kwilteam/kwil-db/pkg/config"

	"github.com/cstockton/go-conv"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	EnvPrefix = "KWILD"
)

type KwildConfig struct {
	GrpcListenAddress   string
	HttpListenAddress   string
	PrivateKey          *ecdsa.PrivateKey
	Deposits            DepositsConfig
	ChainSyncer         ChainSyncerConfig
	WithoutChainSyncer  bool
	WithoutAccountStore bool
	SqliteFilePath      string
	Log                 log.Config
	ExtensionEndpoints  []string
}

type DepositsConfig struct {
	ReconnectionInterval int
	BlockConfirmations   int
	ChainCode            int
	ClientChainRPCURL    string
	PoolAddress          string
}

type ChainSyncerConfig struct {
	ChunkSize int
}

var (
	RegisteredVariables = []config.CfgVar{
		PrivateKey,
		GrpcListenAddress,
		DepositsReconnectionInterval,
		DepositsBlockConfirmation,
		DepositsChainCode,
		DepositsClientChainRPCURL,
		WithoutAccountStore,
		WithoutChainSyncer,
		DepositsPoolAddress,
		ChainSyncerChunkSize,
		SqliteFilePath,
		LogLevel,
		LogOutputPaths,
		HttpListenAddress,
		ExtensionEndpoints,
	}
)

var (
	PrivateKey = config.CfgVar{
		EnvName: "PRIVATE_KEY",
		Field:   "PrivateKey",
		Setter: func(val any) (any, error) {
			if val == nil {
				fmt.Println("no private key provided, generating a new one...")
				return crypto.GenerateKey()
			}

			strVal, err := conv.String(val)
			if err != nil {
				return nil, err
			}

			return crypto.HexToECDSA(strVal)
		},
	}

	GrpcListenAddress = config.CfgVar{
		EnvName: "GRPC_LISTEN_ADDRESS",
		Field:   "GrpcListenAddress",
		Default: ":50051",
	}

	DepositsReconnectionInterval = config.CfgVar{
		EnvName: "DEPOSITS_RECONNECTION_INTERVAL",
		Field:   "Deposits.ReconnectionInterval",
		Default: 30,
	}

	DepositsBlockConfirmation = config.CfgVar{
		EnvName: "DEPOSITS_BLOCK_CONFIRMATIONS",
		Field:   "Deposits.BlockConfirmations",
		Default: 12,
	}

	DepositsChainCode = config.CfgVar{
		EnvName: "DEPOSITS_CHAIN_CODE",
		Field:   "Deposits.ChainCode",
		Default: 0,
	}

	DepositsClientChainRPCURL = config.CfgVar{
		EnvName: "DEPOSITS_CLIENT_CHAIN_RPC_URL",
		Field:   "Deposits.ClientChainRPCURL",
		Default: "http://localhost:8545",
	}

	DepositsPoolAddress = config.CfgVar{
		EnvName: "DEPOSITS_POOL_ADDRESS",
		Field:   "Deposits.PoolAddress",
		Default: "0x0000000000000000000000000000000000000000",
	}

	ChainSyncerChunkSize = config.CfgVar{
		EnvName: "CHAIN_SYNCER_CHUNK_SIZE",
		Field:   "ChainSyncer.ChunkSize",
		Default: 100000,
	}

	SqliteFilePath = config.CfgVar{
		EnvName: "SQLITE_FILE_PATH",
		Field:   "SqliteFilePath",
		Setter: func(val any) (any, error) {
			if val != nil {
				return conv.String(val)
			}

			dirname, err := os.UserHomeDir()
			if err != nil {
				dirname = "/tmp"
			}

			return fmt.Sprintf("%s/.kwil/sqlite/", dirname), nil
		},
	}

	LogLevel = config.CfgVar{
		EnvName: "LOG_LEVEL",
		Field:   "Log.Level",
		Default: "info",
	}

	LogOutputPaths = config.CfgVar{
		EnvName: "LOG_OUTPUT_PATHS",
		Field:   "Log.OutputPaths",
		Setter: func(val any) (any, error) {
			if val == nil {
				return []string{"stdout"}, nil
			}

			str, err := conv.String(val)
			if err != nil {
				return nil, err
			}

			return strings.Split(str, ","), nil
		},
	}

	HttpListenAddress = config.CfgVar{
		EnvName: "HTTP_LISTEN_ADDRESS",
		Field:   "HttpListenAddress",
		Default: ":8080",
	}

	ExtensionEndpoints = config.CfgVar{
		EnvName: "EXTENSION_ENDPOINTS",
		Field:   "ExtensionEndpoints",
		Setter: func(val any) (any, error) {
			if val == nil {
				return nil, nil
			}

			str, err := conv.String(val)
			if err != nil {
				return nil, err
			}

			endpointArr := strings.Split(str, ",")
			for i := range endpointArr {
				endpointArr[i] = strings.TrimSpace(endpointArr[i])
			}

			return endpointArr, nil
		},
	}

	WithoutAccountStore = config.CfgVar{
		EnvName: "WITHOUT_ACCOUNT_STORE",
		Field:   "WithoutAccountStore",
		Default: false,
	}

	WithoutChainSyncer = config.CfgVar{
		EnvName: "WITHOUT_CHAIN_SYNCER",
		Field:   "WithoutChainSyncer",
		Default: false,
	}
)
