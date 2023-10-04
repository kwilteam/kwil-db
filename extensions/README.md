# Extensions

Extensions are compile-time loaded pieces of code that impact core `kwild` functionality. Typically, extensions impact core consensus code, and therefore great care should be taken when implementing and choosing to use certain extensions.

## Interfaces and Drivers

Extensions can made by implementing a driver for one of many interfaces. These implementations should be registered using Go's `init()` function, which will register the driver when the package is loaded.  This is conceptually similar to Go's `database/sql` package, where users can implement custom `database/sql/driver/Driver` implementations.

## Build Tags

To include an extension in a build, users should use [Go's build tags](<https://www.digitalocean.com/community/tutorials/customizing-go-binaries-with-build-tags>). Users can specify what extensions they include by including their respective tags:

### Tag Naming

While you can give any name to your extension's tag, this repo adopts the best practice of prefixing the type of extension with the rest of the name. For example, if we were adding an extension that added standard RSA signatures for authentication, we might name the build tag `auth_rsa`.  We could then include this by running:

```bash
go build -tags auth_rsa
```

Additionally, the build tag `ext_test` is added if the extension should be included as a part of `kwild`'s automated testing.
