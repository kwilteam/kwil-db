package execution

// ExtensionInitializer initializes a new instance of an extension.
// It is called when a Kuneiform schema is deployed that calls "use <extension> {key: "value"} as <name>".
// It is also called when the node starts up, if a database is already deployed that uses the extension.
// The key/value pairs are passed as the metadata parameter.
// When initialize is called, the dataset is not yet accessible.
type ExtensionInitializer func(ctx *DeploymentContext, metadata map[string]string) (ExtensionNamespace, error)

// ExtensionNamespace is a named initialized instance of an extension.
// When a Kuneiform schema calls "use <extension> as <name>", a new instance is
// created for that extension, and is accessible via <name>.
// Instances exist for the lifetime of the deployed dataset, and a single
// dataset can have multiple instances of the same extension.
type ExtensionNamespace interface {
	// Call executes the requested method of the extension.
	// It is up to the extension instance implementation to determine
	// if a method is valid, and to subsequently decode the arguments.
	// The arguments passed in as args, as well as returned, are scalar values.
	Call(scoper *ProcedureContext, method string, inputs []any) ([]any, error)
}
