package kwild

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/app/kwild/cmd/server"
	"github.com/kwilteam/kwil-db/internal/app/kwild/cmd/utils"
	"github.com/kwilteam/kwil-db/internal/app/kwild/cmd/validator"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:               "kwild",
	Short:             "kwild command line interface",
	Long:              "kwild allows you to configure Kwild services",
	SilenceUsage:      true,
	DisableAutoGenTag: true,
}

var kwildCfg = config.DefaultConfig()
var rootDir string

func Execute() error {
	rootCmd.AddCommand(
		validator.NewCmdValidator(kwildCfg),
		server.NewServerCmd(kwildCfg),
		utils.NewCmdGenerator(),
	)
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rootDir, "root_dir", "~/.kwild", "kwild root directory for config and data")
	rootCmd.PersistentPreRunE = extractKwildConfig
}

func extractKwildConfig(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())

	// skip loading config if the parent command has the annotation
	if val, ok := cmd.Parent().Annotations["skip_load_config"]; ok {
		if val == "true" {
			return nil
		}
	}

	rootDir, err := config.ExpandPath(rootDir)
	if err != nil {
		fmt.Println("Error while getting absolute path for config file: ", err)
		return err
	}

	err = kwildCfg.LoadKwildConfig(rootDir)
	if err != nil {
		fmt.Println("Failed to load config: ", err)
		return err
	}
	return nil
}
