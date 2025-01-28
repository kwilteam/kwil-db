package node

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/custom"
	"github.com/kwilteam/kwil-db/app/node/conf"
	"github.com/kwilteam/kwil-db/app/shared/bind"
	"github.com/kwilteam/kwil-db/version"
)

const (
	emptyBlockTimeoutFlag = "consensus.empty-block-timeout"
)

func StartCmd() *cobra.Command {
	var autogen bool
	var dbOwner string
	cmd := &cobra.Command{
		Use:               "start",
		Short:             "Start the node",
		Long:              "The `start` command starts the Kwil DB blockchain node.",
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Args:    cobra.NoArgs,
		Version: version.KwilVersion,
		Example: custom.BinaryConfig.NodeCmd + " start -r .testnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, err := bind.RootDir(cmd)
			if err != nil {
				return err // the parent command needs to set a persistent flag named "root"
			}

			cfg := conf.ActiveConfig()

			bind.Debugf("effective node config (toml):\n%s", bind.LazyPrinter(func() string {
				rawToml, err := cfg.ToTOML()
				if err != nil {
					return fmt.Errorf("failed to marshal config to toml: %w", err).Error()
				}
				return string(rawToml)
			}))

			stopProfiler, err := startProfilers(profMode(cfg.ProfileMode), cfg.ProfileFile)
			if err != nil {
				cmd.Usage()
				return err
			}
			defer stopProfiler()

			// Set the empty block timeout to the propose timeout if not set
			// if the node is running in autogen mode
			if !cmd.Flags().Changed(emptyBlockTimeoutFlag) && autogen {
				cfg.Consensus.EmptyBlockTimeout = cfg.Consensus.ProposeTimeout
			}

			return runNode(cmd.Context(), rootDir, cfg, autogen, dbOwner)
		},
	}

	// Other node flags have config file and env analogs, and will be loaded
	// into koanf where the values are merged.
	defaultCfg := custom.DefaultConfig() // not config.DefaultConfig(), so custom command config is used
	bind.SetFlagsFromStruct(cmd.Flags(), defaultCfg)

	cmd.SetVersionTemplate(custom.BinaryConfig.NodeCmd + " {{printf \"version %s\" .Version}}\n")
	cmd.Flags().BoolVarP(&autogen, "autogen", "a", false,
		"auto generate private key, genesis file, and config file if not exist")
	cmd.Flags().StringVarP(&dbOwner, "db-owner", "d", "", "owner of the database. This is either a hex pubkey or an address string")

	return cmd
}
