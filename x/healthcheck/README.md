## healthcheck check the provided checks.

### Usage

```go
registrar := healthcheck.NewRegistrar()
registrar.RegisterAsyncCheck(10*time.Second, 5*time.Second, healthcheck.Check{
		Name: "dummy",
		Check: func(ctx context.Context) error {
			// error make this check fail, nil will make it succeed
			return nil
		},
})
ck := registrar.BuildChecker(simple_checker.New())
```