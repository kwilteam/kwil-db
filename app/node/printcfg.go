package node

import (
	"fmt"

	"github.com/kwilteam/kwil-db/app/custom"
	"github.com/kwilteam/kwil-db/app/node/conf"
	"github.com/kwilteam/kwil-db/app/shared/bind"

	"github.com/spf13/cobra"
)

func PrintConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print-config",
		Short: "Print the node configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := bind.RootDir(cmd)
			if err != nil {
				return err // the parent command needs to set a persistent flag named "root"
			}

			cfg := conf.ActiveConfig()

			rawToml, err := cfg.ToTOML()
			if err != nil {
				return fmt.Errorf("failed to marshal config to toml: %w", err)
			}

			fmt.Println(string(rawToml))

			return nil
		},
	}

	defaultCfg := custom.DefaultConfig() // not config.DefaultConfig(), so custom command config is used
	bind.SetFlagsFromStruct(cmd.Flags(), defaultCfg)

	return cmd
}
