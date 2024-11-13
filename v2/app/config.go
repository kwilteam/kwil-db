package app

import (
	"fmt"

	"kwil/config"

	"github.com/knadh/koanf/v2"
	gotoml "github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

const (
	ConfigFileName  = "kwil.toml"
	GenesisFileName = "genesis.json"
)

func PrintConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print-config",
		Short: "Print the node configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := RootDir(cmd)
			if err != nil {
				return err // the parent command needs to set a persistent flag named "root"
			}

			// k => config.Config
			var cfg config.Config
			err = k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"})
			if err != nil {
				return fmt.Errorf("failed to unmarshal config: %w", err)
			}

			rawToml, err := gotoml.Marshal(&cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config to toml: %w", err)
			}

			fmt.Println(string(rawToml))

			return nil
		},
	}

	// SetNodeFlags(cmd)
	defaultCfg := config.DefaultConfig()
	SetNodeFlagsFromStruct(cmd, defaultCfg)

	return cmd
}
