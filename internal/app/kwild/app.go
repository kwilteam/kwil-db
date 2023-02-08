package kwild

import (
	"kwil/internal/app/kwild/cmd"
)

func Execute() error {
	cmd.RootCmd.SilenceUsage = true
	return cmd.RootCmd.Execute()
}
