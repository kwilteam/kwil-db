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

func Execute() error {
	rootCmd.AddCommand(
		validator.NewCmdValidator(kwildCfg),
		server.NewServerCmd(kwildCfg),
		utils.NewCmdGenerator(),
	)
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&kwildCfg.RootDir, "home", "", "kwild home directory")
	rootCmd.PersistentPreRunE = extractKwildConfig
}

func extractKwildConfig(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())
	err := kwildCfg.LoadKwildConfig()
	if err != nil {
		fmt.Println("Failed to load config: ", err)
		return err
	}
	return nil
}
