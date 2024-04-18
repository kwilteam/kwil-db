# Kwil Go Client

This folder contains the Go language client for interacting with a Kwil RPC
provider. Package `client` may be used to build a third-party application with
the ability to:

- Retrieve the status of a Kwil network.
- List and retrieve Kuneiform schemas deployed on a Kwil network.
- Deploy and drop schemas.
- Execute mutative actions defined in a schema.
- Call read-only actions without a network transaction.
- Run ad-hoc SQL queries.
- Retrieve account information, such as balance and nonce.
- Check the status and execution outcome of a network transaction.

The `client` package is used by the `kwil-cli` application to provide these
functions on the command line. Go applications may use the package directly.

## Get the `core` Go Module

The `client` package is part of the `core` Go sub-module of the `kwil-db` repository. To use the package in your Go application, add it as a `require` in your project's `go.mod`:

```sh
$ go get github.com/kwilteam/kwil-db/core
go: downloading github.com/kwilteam/kwil-db/core v0.1.2
go: downloading github.com/kwilteam/kwil-db v0.7.2
go: added github.com/kwilteam/kwil-db/core v0.1.2
```

If you did not already have a `go.mod` for your project, create one with `go mod init mykwilapp`, replacing `mykwilapp` with the module name for your project, which is typically a remote git repository location.

Alternatively, can also manually edit your `go.mod` and then run `go mod tidy`.

Your `go.mod` should be similar to the following:

```go
module mykwilapp

go 1.22

require (
    github.com/kwilteam/kwil-db/core v0.1.2
)
```

## Import the `client` package

With the Kwil `core` module added to your `go.mod`, you can use the `client` package in your code by importing it:

```go
import "github.com/kwilteam/kwil-db/core/client"
```

## Using the `Client` type

### Basic functionality

The main functionality is provided by the `Client` type. The `NewClient` function constructs a new `Client` instance from the URL of a Kwil RPC provider, and a set of options in the `core/types/client.Options` type.

For example:

```go
package main

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/client"
	klog "github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	ctypes "github.com/kwilteam/kwil-db/core/types/client"
)

const (
	provider = "https://longhorn.kwil.com"
)

func main() {
	ctx := context.Background()

	// Create the client and connect to the RPC provider.
	cl, err := client.NewClient(ctx, provider, ctypes.DefaultOptions())
	if err != nil {
		log.Fatal(err)
	}

	// Report the chain ID and block height of the provider.
	chainInfo, err := cl.ChainInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connected to Kwil chain %q, block height %d\n",
		chainInfo.ChainID, chainInfo.BlockHeight)
}
```

### Wallet Setup

In the above example, we used `ctypes.DefaultOptions()`, which includes no
logger, signer (wallet), or expected chain ID. To work with a Kwil account, use
the `crypto` and `crypto/auth` packages to create and load private keys. For
example, we can create and load a secp256k1 private key plus an Ethereum
"personal" signer with the following functions:

```go
import (
   	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
)

func genKey() *crypto.Secp256k1PrivateKey {
	key, _ := crypto.GenerateSecp256k1Key()
	return key // fmt.Println(key.Hex())
}

func makeSigner(keyHex string) auth.Signer {
	key, err := crypto.Secp256k1PrivateKeyFromHex(keyHex)
	if err != nil {
		panic(fmt.Sprintf("bad private key: %v", err))
	}
	return &auth.EthPersonalSigner{Key: *key}
}
```

Now we can expand our example application to work with our account and create signed transactions on the specified Kwil network.

```go
const (
	chainID  = "longhorn" // expect provider to report this chain ID
	provider = "https://longhorn.kwil.com"
	privKey  = "..." // my secp256k1 private key in hexadecimal
)

func main() {
	ctx := context.Background()
	signer := makeSigner(privKey)
	acctID := signer.Identity()

	opts := &ctypes.Options{
		Logger:  klog.NewStdOut(klog.InfoLevel),
		ChainID: chainID, // ensure the provider matches
		Signer:  signer,  // required for transactions and auth
	}

	// Create the client and connect to the RPC provider.
	cl, err := client.NewClient(ctx, provider, opts)
	if err != nil {
		log.Fatal(err)
	}

	// Check our account's balance.
	acctInfo, err := cl.GetAccount(ctx, acctID, types.AccountStatusLatest)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Account %x balance = %v, nonce = %d\n", acctID, acctInfo.Balance, acctInfo.Nonce)
}
```

### List Deployed Databases

List any existing databases with the `ListDatabases` method:

```go
// List previously deployed database owned by us.
datasets, err := cl.ListDatabases(ctx, acctID)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Found %d database(s) owned by me.\n", len(datasets))
```

Initially there will be no owned databases. To deploy one, the account will need
to be funded. If the account has no balance, use the
[faucet](https://faucet.kwil.com/) to request testnet tokens. See [Manual Faucet
Use](#manual-faucet-use) to request funds for an address with no web wallet.

### Databases Deployment

Now that we have a `Client` with a working RPC provider connection and a funded
account, we can deploy and drop databases. To deploy one, use the
`DeployDatabase` method. Unlike the methods we have used so far, this one will
create, sign, and broadcast a blockchain transaction on the Kwil network. Once
the transaction is included in a block and executed, the database will become
available for use.

Before deploying a database, the schema definition is required. This is modeled
by the `core/types/transactions.Schema` type. We can parse a Kuneiform `.kf`
file using `github.com/kwilteam/kuneiform/kfparser` as follows:

```go
import "github.com/kwilteam/kuneiform/kfparser"

// unmarshalKf parses the contents of a Kuneiform schema file.
func unmarshalKf(content string) (*transactions.Schema, error) {
	astSchema, err := kfparser.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}
	schemaJSON, err := astSchema.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	var db transactions.Schema
	return &db, json.Unmarshal(schemaJSON, &db)
}
```

In our example app, we can now use `DeployDatabase`:

```go
// Use the kuneiform packages to load the schema.
schema, err := unmarshalKf(testKf)
if err != nil {
	log.Fatal(err)
}

txHash, err := cl.DeployDatabase(ctx, schema)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("DeployDatabase succeeded! txHash = %x", txHash)
```

If that succeeded, the database deployment transaction was successfully
broadcasted. The `txHash` is this transaction's identifier. However, the
database is **not yet deployed!**

First we have to wait for the next block for the transaction to be executed. To
ensure that the transaction is included in a block before the method returns, we
may specify an option as follows:

```go
// When broadcasting a transaction, wait until it is included in a block.
txOpts := []ctypes.TxOpt{ctypes.WithSyncBroadcast(true)}
txHash, err := cl.DeployDatabase(ctx, schema, txOpts...) // wait
```

In addition to broadcasting the transaction, this ensures the transaction was
included in a block. The next section describes how to check the outcome of a
transaction's *execution*.

### Transaction Status

After being broadcasted to a Kwil node, a transaction must be included in a
block *and* execute without error for the database to actually be deployed. Use
the `TxQuery` method to check.

In our example app, we can define the following closure to use after every
transaction we broadcast:

```go
// After broadcast, we get a transaction hash that uniquely identifies the
// transaction. Use the TxQuery method to verify the execution succeeded.
checkTx := func(txHash []byte, desc string) {
	res, err := cl.TxQuery(ctx, txHash)
	if err != nil {
		log.Fatal(err)
	}
	if res.TxResult.Code == transactions.CodeOk.Uint32() {
		fmt.Printf("Success: %q in transaction %x\n", desc, txHash)
	} else {
		log.Fatalf("Fail: %q in transaction %x, Result code %d, log: %q",
			desc, txHash, res.TxResult.Code, res.TxResult.Log)
	}
}

txHash, err := cl.DeployDatabase(ctx, schema, txOpts...)
if err != nil {
	log.Fatal(err)
}
checkTx(txHash, "deploy database")
```

Combined the `WithSyncBroadcast` option in `txOpts`, this use of `TxQuery` will
ensure the transaction executed without error, otherwise the application will exit.

### Dropping a Database

With a successfully deployed database, use the `DropDatabase` method to delete a database:

```go
txHash, err = cl.DropDatabase(ctx, dbName, txOpts...)
if err != nil {
	log.Fatal(err)
}
checkTx(txHash, "drop database")
```

NOTE: This is only permitted if you are the `owner` of the database i.e. you deployed it.

### Action Execution

As with the database deploy and drop methods, action *execution* requires a
transaction since it is used to modify data in a schema.  Use the
`ExecuteAction` method.

In the schema deployed by our example app, the action called `"tag"` will insert data:

```go
const actionName = "tag"
args := [][]any{{"jon was here"}} // one execution, one argument
txHash, err := cl.ExecuteAction(ctx, dbid, actionName, args, txOpts...)
if err != nil {
	log.Fatal(err)
}
checkTx(txHash, "execute action")
```

In the above example, the schema's `"tag"` action is defined as:

```js
action tag($msg) public {
    INSERT INTO "tags" (ident, msg) VALUES (@caller, $msg);
}
```

We called the `"tag"` action with arguments `[][]any{{"jon was here"}}`. This is
a `[][]any` to support batched action execution. In this example, we execute the
action once, using `"jon was here"` as the `$msg` argument.

For example, a batch of two executions of an action that requires three inputs,
such as `action multi_tag($msg1, $msg2, $msg3) public`, might look like:

```go
args := [][]any{
	{"first1", "first2", "first3"},    // first execution
	{"second1", "second2", "second3"}, // second execution
}
```

NOTE: To execute an action with no input arguments, provide `nil`.

### View (read-only) Action Calls

To run a read-only action, which is defined with the `view` modifier, use the
`CallAction` method.

For example, to call the `get_all` method that returns all records in the `tags` table:

```go
// Use a read-only view call (no blockchain transaction) to list all entries
records, err := cl.CallAction(ctx, dbid, "get_all", nil)
if err != nil {
	log.Fatal(err)
}
```

The `CallAction` method returns the data in the `core/types/client.Records`
type. See the [godocs](https://pkg.go.dev/github.com/kwilteam/kwil-db/core/types/client)
for this type to see the methods available for accessing the records.

## Complete Example

For a complete example with the schema used in the sections above, see the code
in [`core/client/example`](./example/main.go).

The `kwil-cli` CLI app is also built on the `Client` type, and
[its code](https://github.com/kwilteam/kwil-db/tree/main/cmd/kwil-cli)
can be used as a reference.

## Manual Faucet Use

If you have an address with no corresponding web3 wallet to connect to the
faucet web page, you can directly request funds with an HTTP POST request. For
example, if the account ID of your generated key from the example app is
"e52f339994377968b5ef84a04f60756ec249734d", you can use `curl` as follows:

```sh
$ curl -X POST --data '{"address": "0xe52f339994377968b5ef84a04f60756ec249734d"}' \
  --header 'content-type: application/json' https://kwil-faucet-server.onrender.com/funds
{"message":"Successfully sent 10 tokens to 0xe52f339994377968b5ef84a04f60756ec249734d. New balance: 14"}
```
