package node

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/custom"
	"github.com/kwilteam/kwil-db/app/shared"
	"github.com/kwilteam/kwil-db/version"
)

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "start",
		Short:             "start the node (default command)",
		Long:              "Start the node running",
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: version.KwilVersion,
		Example: custom.BinaryConfig.NodeCmd + " start -r .testnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, err := shared.RootDir(cmd)
			if err != nil {
				return err // the parent command needs to set a persistent flag named "root"
			}

			cfg := shared.ActiveConfig()

			shared.Debugf("effective node config (toml):\n%s", shared.LazyPrinter(func() string {
				rawToml, err := shared.ConfigToTOML(cfg)
				if err != nil {
					return fmt.Errorf("failed to marshal config to toml: %w", err).Error()
				}
				return rawToml
			}))

			return runNode(cmd.Context(), rootDir, cfg)
		},
	}

	// Other node flags have config file and env analogs, and will be loaded
	// into koanf where the values are merged.
	// SetNodeFlags(cmd)
	defaultCfg := shared.DefaultConfig() // not config.DefaultConfig(), so custom command config is used
	shared.SetNodeFlagsFromStruct(cmd, defaultCfg)

	cmd.SetVersionTemplate(custom.BinaryConfig.NodeCmd + " {{printf \"version %s\" .Version}}\n")

	return cmd
}
