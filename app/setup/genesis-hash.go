package setup

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/app/node/conf"
	"github.com/kwilteam/kwil-db/app/shared/bind"
	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/app/snapshot"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/node"
	"github.com/spf13/cobra"
)

var (
	genesisHashLong = `Compute the genesis hash from existing PostgreSQL datasets, and optionally update a ` + "`genesis.json`" + ` file.
It can be configured to connect to Postgres either using the root directory (from which the ` + "`config.toml`" + ` will be read),
or by specifying the connection details directly.

Alternatively, a snapshot file can be provided to compute the genesis hash from a snapshot file instead of the database. The snapshot can either be
a .sql or .sql.gz file.

By default, it will print the genesis hash to stdout. To specify a genesis file to update as well, use the ` + "`--genesis`" + ` flag.`

	genesisHashExample = `# Compute the genesis hash from an existing PostgreSQL database using a connection, and add it to a genesis file
kwild setup genesis-hash --dbname kwild --host "127.0.0.1" --port "5432" --user kwild --genesis "~/.kwild/abci/config/genesis.json"

# Compute the genesis hash from an existing PostgreSQL database using the root directory
kwild setup genesis-hash --genesis "~/.kwild/abci/config/genesis.json" --root-dir "~/.kwild"

# Compute the genesis hash from a snapshot file
kwild setup genesis-hash --snapshot "/path/to/snapshot.sql.gz" --genesis "~/.kwild/abci/config/genesis.json"`
)

func GenesisHashCmd() *cobra.Command {
	var genesisFile, snapshotFile string

	cmd := &cobra.Command{
		Use:     "genesis-hash",
		Short:   "Compute genesis hash from existing PostgreSQL data, and optionally update genesis.json.",
		Long:    genesisHashLong,
		Example: genesisHashExample,
		Args:    cobra.NoArgs,
		// Override the root's PersistentPreRunE to bind only the config file,
		// not the full node flag set.
		PersistentPreRunE: bind.ChainPreRuns(conf.PreRunBindConfigFileStrict[config.Config]), // but not the flags
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("snapshot") && cmd.Flags().Changed(bind.RootFlagName) {
				return display.PrintErr(cmd, errors.New("cannot use both --snapshot and --root-dir"))
			}

			var appHash []byte
			if cmd.Flags().Changed("snapshot") {
				var err error
				appHash, err = appHashFromSnapshotFile(snapshotFile)
				if err != nil {
					return display.PrintErr(cmd, err)
				}
			} else { // create a snapshot first
				dbCfg := conf.ActiveConfig().DB
				pgConf, err := bind.GetPostgresFlags(cmd, &dbCfg)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("failed to get postgres flags: %v", err))
				}

				dir, err := tmpKwilAdminSnapshotDir()
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				// clean up any previous temp admin snapshots
				err = cleanupTmpKwilAdminDir(dir)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				// ensure the temp admin snapshots directory exists
				err = ensureTmpKwilAdminDir(dir)
				if err != nil {
					return display.PrintErr(cmd, err)
				}
				defer cleanupTmpKwilAdminDir(dir) // clean up temp admin snapshots directory on exit after app hash computation

				_, _, genCfg, err := snapshot.PGDump(cmd.Context(), pgConf.DBName, pgConf.User, pgConf.Pass, pgConf.Host, pgConf.Port, dir)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				appHash = genCfg.StateHash
			}

			if genesisFile != "" {
				err := writeAndReturnGenesisHash(genesisFile, appHash)
				if err != nil {
					return display.PrintErr(cmd, err)
				}
			}
			return display.PrintCmd(cmd, &genesisHashRes{
				Hash: base64.StdEncoding.EncodeToString(appHash),
			})
		},
	}

	cmd.Flags().StringVarP(&genesisFile, "genesis", "g", "", "optional path to the genesis file to patch with the computed app hash")
	cmd.Flags().StringVarP(&snapshotFile, "snapshot", "s", "", "optional path to the snapshot file to use for the genesis hash computation")

	bind.BindPostgresFlags(cmd, &conf.ActiveConfig().DB)
	return cmd
}

// appHashFromSnapshotFile computes the app hash from a snapshot file.
func appHashFromSnapshotFile(filePath string) ([]byte, error) {
	filePath, err := node.ExpandPath(filePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var reader io.Reader
	if filepath.Ext(filePath) == ".gz" {
		reader, err = gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
	} else {
		reader = file
	}

	hash := sha256.New()

	_, err = io.Copy(hash, reader)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func writeAndReturnGenesisHash(genesisFile string, appHash []byte) error {
	genesisFile, err := node.ExpandPath(genesisFile)
	if err != nil {
		return err
	}

	cfg, err := config.LoadGenesisConfig(genesisFile)
	if err != nil {
		return err
	}

	cfg.StateHash = appHash

	return cfg.SaveAs(genesisFile)
}

type genesisHashRes struct {
	Hash string `json:"hash"`
}

func (g *genesisHashRes) MarshalJSON() ([]byte, error) {
	type Alias genesisHashRes
	return json.Marshal(*(*Alias)(g))
}

func (g *genesisHashRes) MarshalText() (text []byte, err error) {
	return []byte("App Hash: " + g.Hash), nil
}

// tmpKwilAdminSnapshotDir returns the temporary directory for kwil-admin snapshots.
func tmpKwilAdminSnapshotDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	r, err := node.ExpandPath(filepath.Join(home, ".kwild-snaps-temp"))
	if err != nil {
		return "", err
	}

	return r, nil
}

// ensureTmpKwilAdminDir ensures that the temporary directory for kwil-admin snapshots exists.
func ensureTmpKwilAdminDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
	}
	return err
}

// cleanupTmpKwilAdminDir removes the temporary directory for kwil-admin snapshots.
func cleanupTmpKwilAdminDir(dir string) error {
	if _, err := os.Stat(dir); err == nil {
		err = os.RemoveAll(dir)
		if err != nil {
			return err
		}
	}

	return nil
}
