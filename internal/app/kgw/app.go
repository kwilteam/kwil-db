package kgw

import (
	"kwil/internal/app/kgw/cmd"
)

func Execute() error {
	cmd.RootCmd.SilenceUsage = true
	return cmd.RootCmd.Execute()
}
