package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/app/shared/bind"
	"github.com/kwilteam/kwil-db/node"
	"github.com/kwilteam/kwil-db/node/pg"

	"github.com/spf13/cobra"
)

func ResetCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the blockchain and the application state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, err := bind.RootDir(cmd)
			if err != nil {
				return err // the parent command needs to set a persistent flag named "root"
			}
			rootDir, err = node.ExpandPath(rootDir)
			if err != nil {
				return err
			}
			if _, err := os.Stat(rootDir); os.IsNotExist(err) {
				return fmt.Errorf("root directory %s does not exist", rootDir)
			}

			pgConf, err := bind.GetPostgresFlags(cmd)
			if err != nil {
				return err
			}

			err = resetPGState(cmd.Context(), pgConf)
			if err != nil {
				return err
			}
			fmt.Printf("Postgres state reset. Host: %s; Port: %s; Database: %s\n", pgConf.Host, pgConf.Port, pgConf.DBName)

			if all {
				// remove the blockstore if all is set
				chainDir := filepath.Join(rootDir, "blockstore")
				if err := os.RemoveAll(chainDir); err != nil {
					return err
				}
				fmt.Printf("Blockstore directory removed: %s\n", chainDir)

				// remove rcvd_snaps if exists
				snapDir := filepath.Join(rootDir, "rcvd_snaps")
				if err := os.RemoveAll(snapDir); err != nil {
					return err
				}
				fmt.Printf("Statesync snapshots directory removed: %s\n", snapDir)

				// remove snapshots if exists
				snapDir = filepath.Join(rootDir, "snapshots")
				if err := os.RemoveAll(snapDir); err != nil {
					return err
				}
				fmt.Println("Snapshots directory removed", snapDir)

				// remove the migrations directory
				migrationsDir := filepath.Join(rootDir, "migrations")
				if err := os.RemoveAll(migrationsDir); err != nil {
					return err
				}
				fmt.Println("Migrations directory removed", migrationsDir)

				// remove genesis state file if exists
				genesisFile := filepath.Join(rootDir, "genesis-state.sql.gz")
				os.Remove(genesisFile) // ignore error
				fmt.Println("Genesis state file removed", genesisFile)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "reset all data, if this is not set, only the app state will be reset")
	bind.BindPostgresFlags(cmd)

	return cmd
}

// resetPGState drops and creates the database.
func resetPGState(ctx context.Context, conf *pg.ConnConfig) error {
	dropDB := conf.DBName
	conf.DBName = "postgres"
	defer func() { conf.DBName = dropDB }()

	conn, err := pg.NewPool(ctx, &pg.PoolConfig{
		ConnConfig: *conf,
		MaxConns:   2, // requires 2 connections
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Execute(ctx, "DROP DATABASE "+dropDB)
	if err != nil {
		return err
	}

	_, err = conn.Execute(ctx, "CREATE DATABASE "+dropDB+" OWNER kwild")
	if err != nil {
		return err
	}

	return nil
}
