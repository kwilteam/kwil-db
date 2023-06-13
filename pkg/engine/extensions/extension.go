package extensions

type Extension struct {
	// Name is the name of the extension.
	// It can be changed by the user.
	Name string

	// Path is the path to the extension's directory.
	Path string

	// Cmd is the command to execute when the extension is called.
	// It should be the name of the executable, and it must be in the same directory as the extension's config file.
	Cmd string

	// GlobalFlags are flags that the user can pass to the extension on initialization.
	GlobalFlags map[string]string

	// Functions are the functions the extension provides.
	Functions map[string]*ExtensionFunction
}

// ExtensionConfig is the configuration for an extension.
// It should be kept within the extension's directory.
type ExtensionConfig struct {
	// Name is the name of the extension.
	// This will be used to call the extension from the application.
	// e.g. extension "mypkg" with function "hello" would be called as mypkg.hello()
	Name string `json:"name" yaml:"name"`

	// Command is the command to execute when the extension is called.
	// It should be the name of the executable, and it must be in the same directory as the extension's config file.
	Cmd string `json:"cmd" yaml:"cmd"`

	// GlobalFlags are flags that the user can pass to the extension on initialization.
	// They will be included with every command.  A default should be provided by the extension.
	// e.g. --token-address=<token_address>
	GlobalFlags map[string]string `json:"global_flags" yaml:"global_flags"`

	// Functions is a list of functions the extension provides.
	Functions []*ExtensionFunction `json:"functions" yaml:"functions"`
}

/*
An ExtensionFunction is a configuration for a function provided by an extension.

If in the application, you wanted to call a command with the three user inputs: "Bitcoin", "Satoshi", and "Nakamoto",
the config may look like this:

CLI Command: myapp name-func 'Bitcoin' --first-name='Satoshi' --last-name='Nakamoto'

Config:

	{
		"name": "hello",
		"cmd": "name-func",
		"inputs": [
			"",
			"--first-name",
			"--last-name"
		]
	}

In the app, this would be called like this:
pkgname.hello("Bitcoin", "Satoshi", "Nakamoto")
*/
type ExtensionFunction struct {
	// Name is the name of the function.
	// It is case insensitive.
	Name string `json:"name" yaml:"name"`

	// Subcommands are the subcommands to execute when the function is called.
	// e.g. "name-func" would be passed as []string{"name-func"}
	Subcommands []string `json:"subcommands" yaml:"subcommands"`

	// Inputs is the list of inputs the function accepts.
	// If the input is a string, it is a flag.
	// e.g. "-name" would be used as -name=<user_value>
	// If the inpuit is an empty string, then it is a positional argument.
	Inputs []string `json:"inputs" yaml:"inputs"`

	// Returns is the amount of values the function returns.
	Returns int `json:"returns" yaml:"returns"`
}
