# Original design

What's a driver?

Driver is the actual implementation that could be used to interact with the system under test.

There should be a DriverInterface that all drivers should implement.

For example:

```go
type KwilClientDriver interface {
    DeployDb(ctx context.Context, schema *Schema) TxHash
    TxQuery(ctx context.Context, txHash TxHash) TxQueryResult
}
```

## What is a specification?

A specification describes all the interactions needed and uses a driver to perform corresponding interactions.

For example:

```go
type DeployDatabaseSpec struct {
    Driver KwilClientDriver
}

func NewDeployDatabaseSpec(driver GrpcDriver) *DeployDatabaseSpec {
    return &DeployDatabaseSpec{
        Driver: driver,
    }
}


func (s *DeployDatabaseSpec) DeployDatabase(ctx context.Context, schema *Schema) error {
    // get schema from somewhere
    s.Driver.DeployDb(ctx, schema)
    return nil
}

func (s *DeployDatabaseSpec) TxSuccess(ctx context.Context, txHash TxHash) error {
    s.Driver.TxQuery(ctx, txHash)
    return nil
}
```

## A test case

```go
func TestDeployDatabase(t *testing.T) {
    ctx := context.Background()
    driver := NewKwilClientDriver()
    spec := NewDeployDatabaseSpec(driver)

    schema := &Schema{
        Name: "test",
    }
    err := spec.DeployDatabase(ctx, schema)
    if err != nil {
        t.Fatal(err)
    }

    txHash := TxHash("0x123")
    err = spec.TxSuccess(ctx, txHash)
    if err != nil {
        t.Fatal(err)
    }
}
```

## what we have right now

It gets a bit cumbersome to write different test cases.

For example, in `TestDeployDatabase` and `TestDropDatabase`, in `DropDatabase` we need to deploy a database first, then drop it.

And for acceptance test purposes, a single test case is enough, since we only test the happy path.

## next

I think it's time to follow the original design to write the tests, especially for the integration tests.
