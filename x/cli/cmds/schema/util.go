package schema

import (
	"context"
	"fmt"
	"strings"

	"kwil/x/schemadef/pgschema"
	"kwil/x/schemadef/sqlschema"
	"kwil/x/sql/sqlclient"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func loadRealm(ctx context.Context, p string, exclude []string) (*sqlschema.Realm, error) {
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
		return pgschema.ParseSchemaFiles(path)
	case "postgres":
		client, err := sqlclient.Open(ctx, p)
		if err != nil {
			return nil, err
		}
		defer client.Close()
		opts := &sqlschema.InspectRealmOption{}
		if client.URL.Schema != "" {
			opts.Schemas = append(opts.Schemas, client.URL.Schema)
		}
		return client.InspectRealm(ctx, opts)
	default:
		return nil, fmt.Errorf("unknown schema scheme: %s", scheme)
	}
}

func planSummary(cmd *cobra.Command, plan *sqlschema.Plan) error {
	cmd.Println("Planned Changes:")
	for _, c := range plan.Changes {
		if c.Comment != "" {
			cmd.Println("--", strings.ToUpper(c.Comment[:1])+c.Comment[1:])
		}
		cmd.Println(color.YellowString("   %s", c.Cmd))
	}
	return nil
}
