package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/kwilteam/kwil-db/pkg/config"

	"github.com/cstockton/go-conv"
)

const (
	EnvPrefix = "KWILD"
)

type KwildConfig struct {
	GrpcListenAddress  string
	HttpListenAddress  string
	PrivateKey         *crypto.Ed25519PrivateKey
	SqliteFilePath     string
	Log                log.Config
	ExtensionEndpoints []string
	BcRpcUrl           string
	WithoutGasCosts    bool
	WithoutNonces      bool
	SnapshotConfig     SnapshotConfig
	RootDir            string
}

type SnapshotConfig struct {
	Enabled         bool
	RecurringHeight uint64
	MaxSnapshots    uint64
	SnapshotDir     string
}

var (
	RegisteredVariables = []config.CfgVar{
		PrivateKey,
		GrpcListenAddress,
		SqliteFilePath,
		LogLevel,
		LogOutputPaths,
		HttpListenAddress,
		ExtensionEndpoints,
		CometBftRPCUrl,
		WithoutGasCosts,
		WithoutNonces,
		RootDir,
	}
)

var (
	PrivateKey = config.CfgVar{
		EnvName: "PRIVATE_KEY",
		Field:   "PrivateKey",
		Setter: func(val any) (any, error) {
			if val == nil {
				fmt.Println("no private key provided, generating a new one...")
				return crypto.GenerateEd25519Key()
			}

			strVal, err := conv.String(val)
			if err != nil {
				return nil, err
			}

			return crypto.Ed25519PrivateKeyFromHex(strVal)
		},
	}

	CometBftRPCUrl = config.CfgVar{
		EnvName: "COMETBFT_RPC_URL",
		Field:   "BcRpcUrl",
		Default: "tcp://localhost:26657",
	}

	GrpcListenAddress = config.CfgVar{
		EnvName: "GRPC_LISTEN_ADDRESS",
		Field:   "GrpcListenAddress",
		Default: ":50051",
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

	WithoutGasCosts = config.CfgVar{
		EnvName: "WITHOUT_GAS_COSTS",
		Field:   "WithoutGasCosts",
		Default: true,
	}

	WithoutNonces = config.CfgVar{
		EnvName: "WITHOUT_NONCES",
		Field:   "WithoutNonces",
		Default: false,
	}

	SnapshotEnabled = config.CfgVar{
		EnvName: "SNAPSHOT_ENABLED",
		Field:   "SnapshotConfig.Enabled",
		Default: false,
	}

	SnapshotRecurringHeight = config.CfgVar{
		EnvName: "SNAPSHOT_RECURRING_HEIGHT",
		Field:   "SnapshotConfig.RecurringHeight",
		Default: uint64(10000), // 12-14 hrs at 1 block per 5 seconds speed
	}

	MaxSnapshots = config.CfgVar{
		EnvName: "MAX_SNAPSHOTS",
		Field:   "SnapshotConfig.MaxSnapshots",
		Default: 2,
	}

	SnapshotDir = config.CfgVar{
		EnvName: "SNAPSHOT_DIR",
		Field:   "SnapshotConfig.SnapshotDir",
		Default: "/tmp/kwil/snapshots",
	}

	RootDir = config.CfgVar{
		EnvName: "ROOT_DIR",
		Field:   "RootDir",
		Setter: func(val any) (any, error) {
			if val == nil {
				home, err := os.UserHomeDir()
				if err != nil {
					// if `home` env(depends on OS) is not set, complain
					// we can use '/tmp/.kwil' or '.kwil' in this case, but it's not a good idea
					return "", err
				}
				return filepath.Join(home, ".kwil"), err
			}

			str, err := conv.String(val)
			if err != nil {
				return nil, err
			}

			return filepath.Clean(str), nil
		},
	}
)
