package setup

import (
	"context"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/spf13/cobra"
)

var (
	resetLong = `To delete all of a Kwil node's data files, use the ` + "`" + `reset` + "`" + ` command. If the root directory is not specified, the node's default root directory will be used.

WARNING: This command should not be used on production systems. This should only be used to reset disposable test nodes.`

	resetExample = `# Delete all of a Kwil node's data files
kwil-admin setup reset --root-dir "~/.kwild"`
)

func resetCmd() *cobra.Command {
	var rootDir string
	var force bool

	resetCmd := &cobra.Command{
		Use:     "reset",
		Short:   "To delete all of a Kwil node's data files, use the `reset` command.",
		Long:    resetLong,
		Example: resetExample,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if rootDir == "" {
				if !force {
					return display.PrintErr(cobraCmd, errors.New("unable to remove default home directory without --force or specifying it using the --root-dir flag"))
				}
				rootDir = common.DefaultKwildRoot()
			} else {
				var err error
				rootDir, err = common.ExpandPath(rootDir)
				if err != nil {
					return display.PrintErr(cobraCmd, err)
				}
			}

			fmt.Println("Resetting all data in", rootDir)

			pgConf, err := getPGConnUsingLocalConfig(cobraCmd, rootDir)
			if err != nil {
				return display.PrintErr(cobraCmd, err)
			}

			err = resetPGState(cobraCmd.Context(), pgConf)
			if err != nil {
				return display.PrintErr(cobraCmd, err)
			}

			expandedRoot, err := common.ExpandPath(rootDir)
			if err != nil {
				return display.PrintErr(cobraCmd, err)
			}

			err = config.ResetAll(expandedRoot)
			if err != nil {
				return display.PrintErr(cobraCmd, err)
			}

			return nil
		},
	}

	resetCmd.Flags().StringVarP(&rootDir, "root-dir", "r", "", "root directory of the kwild node")
	resetCmd.Flags().BoolVarP(&force, "force", "f", false, "force removal of default home directory")
	common.BindPostgresFlags(resetCmd)

	// TODO: remove in v0.10
	resetCmd.Flags().StringP("snappath", "p", "", "path to the snapshot directory")
	resetCmd.Flags().MarkDeprecated("snappath", "this value is no longer used")

	return resetCmd
}

// getPGConnUsingLocalConfig gets the postgres connection that should be used, including configurations set by the local config.
// It uses the same precedence as `kwild`, which is (lowest to highest): default, config file, env. flag.
func getPGConnUsingLocalConfig(cmd *cobra.Command, rootDir string) (*pg.ConnConfig, error) {
	cfg := config.EmptyConfig()
	cfg.RootDir = rootDir

	// merge it with any configured values
	cfg, _, err := config.GetCfg(cfg)
	if err != nil {
		return nil, err
	}

	conf := &pg.ConnConfig{
		Host:   cfg.AppConfig.DBHost,
		Port:   cfg.AppConfig.DBPort,
		User:   cfg.AppConfig.DBUser,
		Pass:   cfg.AppConfig.DBPass,
		DBName: cfg.AppConfig.DBName,
	}

	return common.MergePostgresFlags(conf, cmd)
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
