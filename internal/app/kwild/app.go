package kwild

import (
	"github.com/kwilteam/kwil-db/internal/app/kwild/cmd"
)

func Execute() error {
	cmd.RootCmd.SilenceUsage = true
	return cmd.RootCmd.Execute()
}
