package setup

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/snapshot"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/spf13/cobra"
)

//nolint:unused
var (
	genesisHashLong = `Compute the genesis hash from existing PostgreSQL datasets, and optionally update a ` + "`" + `genesis.json` + "`" + ` file.
It can be configured to connect to Postgres either using the root directory (from which the ` + "`" + `config.toml` + "`" + ` will be read),
or by specifying the connection details directly.

Alternatively, a snapshot file can be provided to compute the genesis hash from a snapshot file instead of the database. The snapshot can either be
a .sql or .sql.gz file.

By default, it will print the genesis hash to stdout. To specify a genesis file to update as well, use the ` + "`" + `--genesis` + "`" + ` flag.`

	genesisHashExample = `# Compute the genesis hash from an existing PostgreSQL database using a connection, and add it to a genesis file
kwil-admin setup genesis-hash --dbname kwild --host "127.0.0.1" --port "5432" --user kwild --genesis "~/.kwild/abci/config/genesis.json"

# Compute the genesis hash from an existing PostgreSQL database using the root directory
kwil-admin setup genesis-hash --genesis "~/.kwild/abci/config/genesis.json" --root-dir "~/.kwild"

# Compute the genesis hash from a snapshot file
kwil-admin setup genesis-hash --snapshot "/path/to/snapshot.sql.gz" --genesis "~/.kwild/abci/config/genesis.json"`
)

func genesisHashCmd() *cobra.Command { //nolint:unused
	var genesisFile string
	var rootDir string
	var snapshotFile string

	cmd := &cobra.Command{
		Use:     "genesis-hash",
		Short:   "Compute genesis hash from existing PostgreSQL datasets, and optionally update `genesis.json`.",
		Long:    genesisHashLong,
		Example: genesisHashExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("snapshot") && cmd.Flags().Changed("root-dir") {
				return display.PrintErr(cmd, errors.New("cannot use both --snapshot and --root-dir"))
			}

			var appHash []byte
			if cmd.Flags().Changed("snapshot") {
				var err error
				appHash, err = appHashFromSnapshotFile(snapshotFile)
				if err != nil {
					return display.PrintErr(cmd, err)
				}
			} else {

				var pgConf *pg.ConnConfig
				var err error
				if rootDir != "" {
					rootDir, err = common.ExpandPath(rootDir)
					if err != nil {
						return display.PrintErr(cmd, err)
					}

					pgConf, err = getPGConnUsingLocalConfig(cmd, rootDir)
					if err != nil {
						return display.PrintErr(cmd, err)
					}
				}

				// clean up any previous temp admin snapshots
				err = cleanupTmpKwilAdminDir()
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				// ensure the temp admin snapshots directory exists
				err = ensureTmpKwilAdminDir()
				if err != nil {
					return display.PrintErr(cmd, err)
				}
				defer cleanupTmpKwilAdminDir()

				dir, err := tmpKwilAdminSnapshotDir()
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				// TODO: once https://github.com/kwilteam/kwil-db/pull/985 goes in, we can delete the max row size.
				_, _, genCfg, err := snapshot.PGDump(cmd.Context(), pgConf.DBName, pgConf.User, pgConf.Pass, pgConf.Host, pgConf.Port, dir)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				appHash = genCfg.DataAppHash
			}

			return writeAndReturnGenesisHash(cmd, genesisFile, appHash)
		},
	}

	cmd.Flags().StringVarP(&genesisFile, "genesis", "g", "", "optional path to the genesis file to patch with the computed app hash")
	cmd.Flags().StringVarP(&rootDir, "root-dir", "r", "", "optional path to the root directory of the kwild node from which the genesis hash will be computed")
	cmd.Flags().StringVarP(&snapshotFile, "snapshot", "s", "", "optional path to the snapshot file to use for the genesis hash computation")
	common.BindPostgresFlags(cmd)

	return cmd
}

func writeAndReturnGenesisHash(cmd *cobra.Command, genesisFile string, appHash []byte) error {
	if genesisFile != "" {
		genesisFile, err := common.ExpandPath(genesisFile)
		if err != nil {
			return display.PrintErr(cmd, err)
		}

		file, err := os.ReadFile(genesisFile)
		if err != nil {
			return display.PrintErr(cmd, err)
		}

		var cfg chain.GenesisConfig
		err = json.Unmarshal(file, &cfg)
		if err != nil {
			return display.PrintErr(cmd, err)
		}

		cfg.DataAppHash = appHash

		file, err = json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return display.PrintErr(cmd, err)
		}

		err = os.WriteFile(genesisFile, file, 0644)
		if err != nil {
			return display.PrintErr(cmd, err)
		}
	}

	return display.PrintCmd(cmd, &genesisHashRes{
		Hash: base64.StdEncoding.EncodeToString(appHash),
	})
}

// appHashFromSnapshotFile computes the app hash from a snapshot file.
func appHashFromSnapshotFile(filePath string) ([]byte, error) {
	filePath, err := common.ExpandPath(filePath)
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

type genesisHashRes struct {
	Hash string
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

	r, err := common.ExpandPath(filepath.Join(home, ".kwil-admin-snaps-temp"))
	if err != nil {
		return "", err
	}

	return r, nil
}

// ensureTmpKwilAdminDir ensures that the temporary directory for kwil-admin snapshots exists.
func ensureTmpKwilAdminDir() error {
	dir, err := tmpKwilAdminSnapshotDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

// cleanupTmpKwilAdminDir removes the temporary directory for kwil-admin snapshots.
func cleanupTmpKwilAdminDir() error {
	dir, err := tmpKwilAdminSnapshotDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(dir); err == nil {
		err = os.RemoveAll(dir)
		if err != nil {
			return err
		}
	}

	return nil
}
