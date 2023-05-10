package config

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/log"
	"os"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/config"

	"github.com/cstockton/go-conv"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	EnvPrefix = "KWILD"
)

type KwildConfig struct {
	GrpcListenAddress string
	HttpListenAddress string
	PrivateKey        *ecdsa.PrivateKey
	Deposits          DepositsConfig
	SqliteFilePath    string
	Log               log.Config
}

type DepositsConfig struct {
	ReconnectionInterval int
	BlockConfirmations   int
	ChainCode            int
	ClientChainRPCURL    string
	PoolAddress          string
}

var (
	RegisteredVariables = []config.CfgVar{
		PrivateKey,
		GrpcListenAddress,
		DepositsReconnectionInterval,
		DepositsBlockConfirmation,
		DepositsChainCode,
		DepositsClientChainRPCURL,
		DepositsPoolAddress,
		SqliteFilePath,
		LogLevel,
		LogOutputPaths,
		HttpListenAddress,
	}
)

var (
	PrivateKey = config.CfgVar{
		EnvName:  "PRIVATE_KEY",
		Required: true,
		Field:    "PrivateKey",
		Setter: func(val any) (any, error) {
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
		EnvName:  "DEPOSITS_POOL_ADDRESS",
		Field:    "Deposits.PoolAddress",
		Required: true,
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
)
