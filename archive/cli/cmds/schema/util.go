package schema

import (
	"kwil/x/proto/apipb"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func planSummaryProto(cmd *cobra.Command, plan *apipb.Plan) error {
	cmd.Println("Planned Changes:")
	for _, c := range plan.Changes {
		if c.Comment != "" {
			cmd.Println("--", strings.ToUpper(c.Comment[:1])+c.Comment[1:])
		}
		cmd.Println(color.YellowString("   %s", c.Cmd))
	}
	return nil
}
