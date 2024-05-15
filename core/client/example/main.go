package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/big"
	"slices"
	"strings"

	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	klog "github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	ctypes "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"
	"github.com/kwilteam/kwil-db/parse"
)

const (
	chainID  = "longhorn"
	provider = "https://longhorn.kwil.com"

	privKey = "..."
)

func main() {
	ctx := context.Background()
	signer := makeEthSigner(privKey)
	acctID := signer.Identity()

	ctypes.DefaultOptions()
	opts := &ctypes.Options{
		Logger:  klog.NewStdOut(klog.InfoLevel),
		ChainID: chainID,
		Signer:  signer, // required only transactions and auth
	}

	// Create the client and connect to the RPC provider.
	cl, err := client.NewClient(ctx, provider, opts)
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

	// Check our account's balance.
	acctInfo, err := cl.GetAccount(ctx, acctID, types.AccountStatusLatest)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Account %x balance = %v, nonce = %d\n", acctID, acctInfo.Balance, acctInfo.Nonce)

	// List previously deployed database owned by us.
	datasets, err := cl.ListDatabases(ctx, acctID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d database(s) owned by me.\n", len(datasets))

	// When broadcasting a transaction, wait until it is included in a block.
	txOpts := []ctypes.TxOpt{ctypes.WithSyncBroadcast(true)}

	// After broadcast, we get a transaction hash that uniquely identifies the
	// transaction. Use the TxQuery method to get the execution result.
	checkTx := func(txHash []byte, attempt string) {
		res, err := cl.TxQuery(ctx, txHash)
		if err != nil {
			log.Fatal(err)
		}
		if res.TxResult.Code == transactions.CodeOk.Uint32() {
			fmt.Printf("Success: %q in transaction %x\n", attempt, txHash)
		} else {
			log.Fatalf("Fail: %q in transaction %x, Result code %d, log: %q",
				attempt, txHash, res.TxResult.Code, res.TxResult.Log)
		}
	}

	// Deploy a Kuneiform schema called "was_here".
	dbName := "was_here"
	dbid := utils.GenerateDBID(dbName, acctID) // derive DBID

	// See if it already deployed.
	deployed := slices.ContainsFunc(datasets, func(d *types.DatasetIdentifier) bool {
		return d.Name == dbName
	})

	if !deployed { // need to deploy the "was_here" database
		// Use the kuneiform packages to load the schema.
		schema, err := unmarshalKf(strings.NewReader(testKf))
		if err != nil {
			log.Fatal(err)
		}

		const minBal int64 = 1e6 // dust
		if acctInfo.Balance.Cmp(big.NewInt(minBal)) < 0 {
			log.Fatalf("Account lacks sufficient funds to deploy a database.")
		}

		fmt.Printf("Deploying database %v...\n", schema.Name)
		txHash, err := cl.DeployDatabase(ctx, schema, txOpts...)
		if err != nil {
			log.Fatal(err)
		}
		checkTx(txHash, "deploy database")
	} else { // already deployed
		// The tags table allows only one entry per user, so we will delete all
		// entries, which we can do because of the action's "owner" modifier.
		deleteAllAction := "delete_all"
		fmt.Printf("Executing action %q to clear database %q...\n", deleteAllAction, dbName)
		txHash, err := cl.Execute(ctx, dbid, deleteAllAction, nil, ctypes.WithSyncBroadcast(true))
		if err != nil {
			log.Fatal(err)
		}
		checkTx(txHash, "execute action")
	}

	// Insert some data with this schema's "tag" action.
	const tagAction = "tag"
	fmt.Printf("Executing action %q to insert data...\n", tagAction)
	txHash, err := cl.Execute(ctx, dbid, tagAction, [][]any{{"jon was here"}}, txOpts...)
	if err != nil {
		log.Fatal(err)
	}
	checkTx(txHash, "execute action")

	// Use a read-only view call (no blockchain transaction) to list all entries
	const getAllAction = "get_all"
	records, err := cl.Call(ctx, dbid, getAllAction, nil)
	if err != nil {
		log.Fatal(err)
	}
	if tab := records.ExportString(); len(tab) == 0 {
		fmt.Println("No data records in table.")
	} else {
		fmt.Println("All entries in tags table:")
		var headers []string
		for k := range tab[0] {
			headers = append(headers, k)
		}
		fmt.Printf("column names: %#v\n"+"values:\n", headers)
		for _, row := range tab {
			var rowVals []string
			for _, h := range headers {
				rowVals = append(rowVals, row[h])
			}
			fmt.Printf("%#v\n", rowVals)
		}
	}

	fmt.Printf("Dropping database %v...\n", dbName)
	txHash, err = cl.DropDatabase(ctx, dbName, txOpts...)
	if err != nil {
		log.Fatal(err)
	}
	checkTx(txHash, "drop database")
}

func makeEthSigner(keyHex string) auth.Signer {
	key, err := crypto.Secp256k1PrivateKeyFromHex(keyHex) // 32 bytes / 64 hex chars
	if err != nil {
		panic(fmt.Sprintf("bad private key: %v", err))
	}
	return &auth.EthPersonalSigner{Key: *key}
}

func genEthKey() *crypto.Secp256k1PrivateKey {
	key, _ := crypto.GenerateSecp256k1Key()
	return key // fmt.Println(key.Hex())
}

func makeEdSigner(keyHex string) auth.Signer {
	key, err := crypto.Ed25519PrivateKeyFromHex(keyHex) // 64 bytes / 128 hex chars
	if err != nil {
		panic(fmt.Sprintf("bad private key: %v", err))
	}
	return &auth.Ed25519Signer{*key}
}

func genEdKey() *crypto.Ed25519PrivateKey {
	key, _ := crypto.GenerateEd25519Key()
	return key // fmt.Println(key.Hex())
}

var testKf = `database was_here;

table tags {
    ident text primary notnull,
    val int default(42),
    msg text notnull
}

action tag($msg) public {
    INSERT INTO "tags" (ident, msg) VALUES (@caller, $msg);
}

action delete_mine() public {
    DELETE FROM tags WHERE ident = @caller;
}

action delete_other ($ident) public owner {
    DELETE FROM "tags" WHERE ident = $ident;
}

action delete_all () public owner {
    DELETE FROM tags;
}

action get_user_tag($ident) public view {
    SELECT msg, val FROM tags WHERE ident = $ident;
}

action get_my_tag() public view {
    SELECT msg, val FROM tags WHERE ident = @caller;
}

action get_all() public view {
    SELECT * FROM tags;
}
`

// go:embed test.json
// var testJSON []byte

func unmarshalKf(file io.Reader) (*types.Schema, error) {
	source, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read Kuneiform source file: %w", err)
	}

	parseRes, err := parse.ParseKuneiform(string(source))
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	} // kfSchema := astSchema.(*schema.Schema); j, _ := json.Marshal(kfSchema)

	return parseRes.Schema, nil
}
