// package custom allows for the creation of a custom-branded CLI that packages together the kwil-cli, kwil-admin, and kwild CLIs.
package custom

import (
	"fmt"

	"github.com/spf13/cobra"

	// kwilCLIRoot "github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds"
	kwildRoot "github.com/kwilteam/kwil-db/app"
	"github.com/kwilteam/kwil-db/config"
)

// binaryConfig configures the generated binary. It is able to control the binary names.
// It is primarily used for generating useful help commands that have proper names.
type binaryConfig struct {
	// ProjectName is the name of the project, which will be used in the help text.
	ProjectName string
	// RootCmd is the name of the root command.
	// If we are building kwild / kwil-cli, then RootCmd is empty.
	RootCmd string
	// NodeCmd is the name of the node command.
	NodeCmd string
	// ClientCmd is the name of the client command.
	ClientCmd string
}

var BinaryConfig = defaultBinaryConfig()

func (b *binaryConfig) NodeUsage() string {
	if b.RootCmd != "" {
		return b.RootCmd + " " + b.NodeCmd
	}
	return b.NodeCmd
}

func (b *binaryConfig) ClientUsage() string {
	if b.RootCmd != "" {
		return b.RootCmd + " " + b.ClientCmd
	}
	return b.ClientCmd
}

func defaultBinaryConfig() binaryConfig {
	return binaryConfig{
		ProjectName: "Kwil",
		NodeCmd:     "kwild",
		ClientCmd:   "kwil-cli",
	}
}

// CmdConfig configures the root command.
type CmdConfig struct {
	// RootCmd is the name of the command.
	RootCmd string
	// ProjectName is the name of the project, which will be used in the help text.
	ProjectName string
	// DefaultConfig is a function that allows the default configuration to be changed.
	DefaultConfig func(*config.Config)
}

var DefaultConfig = config.DefaultConfig()

var longDesc = `Command line interface client for using %s.

There are 3 subcommands:

- ` + "`" + `node` + "`" + `: Runs the %s node and RPC server.
- ` + "`" + `client` + "`" + `: Command line interface for an RPC client.
- ` + "`" + `admin` + "`" + `: Utilities for managing a %s node and network participation.

For guides and reference documentation, see the following links. The links document ` + "`" + `kwild` + "`" + `,
` + "`" + `kwil-cli` + "`" + `, and ` + "`" + `kwil-admin` + "`" + `, which directly correspond to the ` + "`" + `node` + "`" +
	`, ` + "`" + `client` + "`" + `, and ` + "`" + `admin` + "`" + `
subcommands, respectively:

- Node: https://docs.kwil.com/docs/node/quickstart
- Client: https://docs.kwil.com/docs/ref/kwil-cli
- Admin: https://docs.kwil.com/docs/admin/installation`

func NewCustomCmd(cmdConfig CmdConfig) *cobra.Command {
	BinaryConfig.ProjectName = cmdConfig.ProjectName
	BinaryConfig.ClientCmd = "client"
	BinaryConfig.NodeCmd = "node"
	BinaryConfig.RootCmd = cmdConfig.RootCmd

	root := &cobra.Command{
		Use:   cmdConfig.RootCmd,
		Short: "Command line interface client for using " + cmdConfig.ProjectName + ".",
		Long:  fmt.Sprintf(longDesc, cmdConfig.ProjectName, cmdConfig.ProjectName, cmdConfig.ProjectName),
	}

	if cmdConfig.DefaultConfig == nil {
		cmdConfig.DefaultConfig = func(c *config.Config) {}
	}

	if cmdConfig.ProjectName == "" {
		cmdConfig.ProjectName = "kwil"
	}

	if cmdConfig.RootCmd == "" {
		cmdConfig.RootCmd = cmdConfig.ProjectName
	}

	// DefaultConfig = func() *config.Config {
	// 	cmdConfig.DefaultConfig(config.DefaultConfig())
	// 	return oldDefault
	// }

	root.AddCommand(kwildRoot.RootCmd())
	// root.AddCommand(kwilCLIRoot.NewRootCmd()) // TODO

	return root
}
