package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/kwilteam/kwil-db/internal/schemadef/postgres"
	"github.com/kwilteam/kwil-db/internal/schemadef/schema"
	"github.com/kwilteam/kwil-db/internal/schemadef/sqlclient"
	"github.com/spf13/cobra"
)

func loadSchema(ctx context.Context, p string, exclude []string) (*schema.Schema, error) {
	parts := strings.SplitN(p, "://", 2)
	var scheme, path string
	switch len(parts) {
	case 2:
		scheme, path = parts[0], parts[1]
	case 1:
		scheme, path = "file", parts[0]
	}

	switch scheme {
	case "file":
		return postgres.ParseSchemaFiles(path)
	case "postgres":
		client, err := sqlclient.Open(ctx, p)
		if err != nil {
			return nil, err
		}
		defer client.Close()
		return client.InspectSchema(ctx, client.URL.Schema, &schema.InspectOptions{Exclude: exclude})
	default:
		return nil, fmt.Errorf("unknown schema scheme: %s", scheme)
	}
}

func planSummary(cmd *cobra.Command, plan *schema.Plan) error {
	cmd.Println("Planned Changes:")
	for _, c := range plan.Changes {
		if c.Comment != "" {
			cmd.Println("--", strings.ToUpper(c.Comment[:1])+c.Comment[1:])
		}
		cmd.Println(color.YellowString("   %s", c.Cmd))
	}
	return nil
}
