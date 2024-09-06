package common

// BinaryConfigStruct configures the generated binary. It is able to control the binary names.
// It is primarily used for generating useful help commands that have proper names.
// TODO: unexport
type BinaryConfigStruct struct {
	// RootCmd is the name of the root command.
	// If we are building kwild / kwil-cli / kwil-admin, then
	// RootCmd is empty.
	RootCmd string
	// NodeCmd is the name of the node command.
	NodeCmd string
	// ClientCmd is the name of the client command.
	ClientCmd string
	// AdminCmd is the name of the admin command.
	AdminCmd string
	// ProjectName is the name of the project, which will be used in the help text.
	ProjectName string
}

var BinaryConfig = DefaultBinaryConfig()

func (b *BinaryConfigStruct) NodeUsage() string {
	if b.RootCmd != "" {
		return b.RootCmd + " " + b.NodeCmd
	}
	return b.NodeCmd
}

func (b *BinaryConfigStruct) ClientUsage() string {
	if b.RootCmd != "" {
		return b.RootCmd + " " + b.ClientCmd
	}
	return b.ClientCmd
}

func (b *BinaryConfigStruct) AdminUsage() string {
	if b.RootCmd != "" {
		return b.RootCmd + " " + b.AdminCmd
	}
	return b.AdminCmd
}

func DefaultBinaryConfig() BinaryConfigStruct {
	return BinaryConfigStruct{
		ProjectName: "Kwil",
		NodeCmd:     "kwild",
		ClientCmd:   "kwil-cli",
		AdminCmd:    "kwil-admin",
	}
}
