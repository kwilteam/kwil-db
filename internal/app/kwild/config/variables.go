package config

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/extensions"
	"github.com/kwilteam/kwil-db/pkg/log"

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
	ChainSyncer       ChainSyncerConfig
	SqliteFilePath    string
	Log               log.Config
	Extensions        ExtensionConfigs
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

type ExtensionConfigs struct {
	ConfigFilePath string
	Extensions     []*extensions.ExtensionConfig `json:"extensions"`
}

var (
	RegisteredVariables = []config.CfgVar{
		ExtensionConfigFilePath,
		PrivateKey,
		GrpcListenAddress,
		DepositsReconnectionInterval,
		DepositsBlockConfirmation,
		DepositsChainCode,
		DepositsClientChainRPCURL,
		DepositsPoolAddress,
		ChainSyncerChunkSize,
		SqliteFilePath,
		LogLevel,
		LogOutputPaths,
		HttpListenAddress,
		//Extensions,
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

	// TODO: this is a mess.  Gotta fix this
	ExtensionConfigFilePath = config.CfgVar{
		EnvName: "EXTENSION_CONFIG_FILE_PATH",
		Field:   "Extensions",
		Setter: func(val any) (any, error) {
			if val == nil {
				return nil, nil
			}

			path, err := conv.String(val)
			if err != nil {
				return nil, err
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}

			// read the file
			file, err := os.ReadFile(absPath)
			if err != nil {
				return nil, err
			}

			// unmarshal the file
			var exts ExtensionConfigs
			err = json.Unmarshal(file, &exts.Extensions)
			if err != nil {
				return nil, err
			}

			exts.ConfigFilePath = absPath

			return exts, nil
		},
	}

	/*
		Extensions = config.CfgVar{
			EnvName: "EXTENSIONS",
			Field:   "Extensions",
			Setter: func(val any) (any, error) {
				if val == nil {
					return nil, nil
				}

				bts, err := json.Marshal(val)
				if err != nil {
					return nil, err
				}

				var exts ExtensionConfigs

				err = json.Unmarshal(bts, &exts.Extensions)
				if err != nil {
					return nil, err
				}

				return exts, nil
			},

		}*/
)
