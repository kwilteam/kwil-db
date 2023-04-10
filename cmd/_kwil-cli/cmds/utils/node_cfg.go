package utils

import (
	"context"
	"fmt"
	"html/template"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	grpc "kwil/pkg/grpc/client/v1"
	"os"
	"text/tabwriter"

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
			return runServerCfg(cmd.Context(), &opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.format, "format", "f", "", "Format the output using the given Go template")

	return cmd
}

func runServerCfg(ctx context.Context, opts *cfgOptions) error {
	tmpl := template.New("version")
	// load different template according to the opts.format
	cfgTemplate := cfgYamlTemplate
	tmpl, err := tmpl.Parse(cfgTemplate)
	if err != nil {
		return errors.Wrap(err, "template parsing error")
	}

	clt, err := client.New(ctx, config.Config.Node.KwilProviderRpcUrl,
		client.WithoutServiceConfig(),
	)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return prettyPrint(&grpc.SvcConfig{
		ChainCode:       int64(clt.ChainCode),
		PoolAddress:     clt.PoolAddress,
		ProviderAddress: clt.ProviderAddress,
	}, tmpl)
}

func prettyPrint(svcCfg *grpc.SvcConfig, tmpl *template.Template) error {
	t := tabwriter.NewWriter(os.Stdout, 20, 1, 1, ' ', 0)
	err := tmpl.Execute(t, svcCfg)
	_, _ = t.Write([]byte("\n"))
	t.Flush()
	return err
}
