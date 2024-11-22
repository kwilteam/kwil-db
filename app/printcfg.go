package app

import (
	"fmt"

	"github.com/kwilteam/kwil-db/app/shared"

	gotoml "github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

func PrintConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print-config",
		Short: "Print the node configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := shared.RootDir(cmd)
			if err != nil {
				return err // the parent command needs to set a persistent flag named "root"
			}

			cfg := shared.ActiveConfig()

			rawToml, err := gotoml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config to toml: %w", err)
			}

			fmt.Println(string(rawToml))

			return nil
		},
	}

	// SetNodeFlags(cmd)
	defaultCfg := shared.DefaultConfig() // not config.DefaultConfig(), so custom command config is used
	shared.SetNodeFlagsFromStruct(cmd, defaultCfg)

	return cmd
}
