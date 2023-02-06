package kwil_gateway

import (
	"kwil/internal/app/kwil-gateway/cmd"
)

func Execute() error {
	cmd.RootCmd.SilenceUsage = true
	return cmd.RootCmd.Execute()
}
