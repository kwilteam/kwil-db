# Extensions

Extensions are compile-time-loaded pieces of code that impact core `kwild` functionality. Typically, extensions impact core consensus code, and therefore great care should be taken when implementing and choosing to use certain extensions.

## Interfaces and Drivers

Extensions can be made by implementing a driver for one of many interfaces. These implementations should be registered using Go's `init()` function, which will register the driver when the package is loaded.  This is conceptually similar to Go's `database/sql` package, where users can implement custom `database/sql/driver/Driver` implementations.

## Implementation

The code for an extension may be in one of the packages here, or in an external imported module. However, for the extension to be used by `kwild`, it must be imported by this package so that it is available when the application is built. If the extension code exists in an external module, this is done with a simple .go source file that has a nameless import so that the extension's `init()` function is called.

## Build

Extensions can be included in a build either by dropping a register file in this directory, or by including the file in another directory and using build tags

### Register File

To create a register file, simply create a new file called `my_extension.go` in this directory:

```go
package extension.go

import _ "github.com/my_org/kwil-db/path/to/extension"
```

Any other directory this imports will be included in during compilation.

### Build Tags

To include an extension in a build, users should use [Go's build tags](https://pkg.go.dev/cmd/go#hdr-Build_constraints). Users can specify what extensions they include by including their respective tags:

#### Tag Naming

While you can give any name to your extension's tag, this repo adopts the best practice of prefixing the type of extension with the rest of the name. For example, if we were adding an extension that added standard RSA signatures for authentication, we might name the build tag `auth_rsa`.  We could then include this by running:

```bash
go build -tags auth_rsa
```

Additionally, the build tag `ext_test` is added if the extension should be included as a part of `kwild`'s automated testing.

In the above example, the source file that either implements or imports the extension would have the following at the start of the file:

```go
//go:build auth_rsa
```
