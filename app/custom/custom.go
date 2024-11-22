// package custom allows for the creation of a custom-branded CLI that packages
// together the kwil-cli, kwil-admin, and kwild CLIs.
package custom

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
