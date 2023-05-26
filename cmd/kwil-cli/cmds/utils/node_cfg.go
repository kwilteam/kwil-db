package utils

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"text/tabwriter"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	grpc "github.com/kwilteam/kwil-db/pkg/grpc/client/v1"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var cfgYamlTemplate = `
Funding:
  ChainCode:	{{.Funding.ChainCode}}
  PoolAddress:	{{.Funding.PoolAddress}}
  ProviderAddress:	{{.Funding.ProviderAddress}}
  RpcUrl:	{{.Funding.RpcUrl}}
Gateway:
  GraphqlUrl:	{{.Gateway.GraphqlUrl}}
`

type cfgOptions struct {
	format string
}

func NewServerCfgCmd() *cobra.Command {
	var opts cfgOptions

	var cmd = &cobra.Command{
		Use:   "node-config [OPTIONS]",
		Short: "Show connected node configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutPrivateKey, func(ctx context.Context, client *client.Client, config *config.KwilCliConfig) error {
				tmpl := template.New("version")
				// load different template according to the opts.format
				cfgTemplate := cfgYamlTemplate
				tmpl, err := tmpl.Parse(cfgTemplate)
				if err != nil {
					return errors.Wrap(err, "template parsing error")
				}

				cfg, err := client.GetConfig(ctx)
				if err != nil {
					return errors.Wrap(err, "error getting node configuration")
				}

				printCfg(cfg)
				return nil
			})
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.format, "format", "f", "", "Format the output using the given Go template")

	return cmd
}

func printCfg(cfg *grpc.SvcConfig) {
	fmt.Printf("ChainCode: %d\n", cfg.ChainCode)
	fmt.Printf("PoolAddress: %s\n", cfg.PoolAddress)
	fmt.Printf("ProviderAddress: %s\n", cfg.ProviderAddress)
}

func prettyPrint(svcCfg *grpc.SvcConfig, tmpl *template.Template) error {
	t := tabwriter.NewWriter(os.Stdout, 20, 1, 1, ' ', 0)
	err := tmpl.Execute(t, svcCfg)
	_, _ = t.Write([]byte("\n"))
	t.Flush()
	return err
}
