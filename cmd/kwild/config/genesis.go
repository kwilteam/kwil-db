package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
)

const (
	nodeDirPerm = 0755
)

// loadGenesisAndPrivateKey generates private key and genesis file if not exist
//
//   - If genesis file exists but not private key file, it will generate private
//     key and start the node as a non-validator.
//   - Otherwise, the genesis file is generated based on the private key and
//     starts the node as a validator.
func loadGenesisAndPrivateKey(autoGen bool, privKeyPath, rootDir string) (privKey *crypto.Ed25519PrivateKey, genesisCfg *chain.GenesisConfig, err error) {
	// Get private key:
	//  - if private key file exists, load it.
	//  - else if in autogen mode, generate private key and write to file.
	//  - else fail

	if err = os.MkdirAll(rootDir, nodeDirPerm); err != nil {
		return nil, nil, fmt.Errorf("failed to make root directory: %w", err)
	}

	chainRootDir := filepath.Join(rootDir, ABCIDirName)
	priv, pub, newKey, err := ReadOrCreatePrivateKeyFile(privKeyPath, autoGen)
	if err != nil {
		return nil, nil, err
	}
	if newKey {
		fmt.Printf("Generated new private key, path: %v\n", privKeyPath)
	}
	privKey, err = crypto.Ed25519PrivateKeyFromBytes(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid private key: %v", err)
	}

	genFile := filepath.Join(rootDir, cometbft.GenesisJSONName)

	if fileExists(genFile) {
		genesisCfg, err = chain.LoadGenesisConfig(genFile)
		if err != nil {
			return nil, nil, fmt.Errorf("error loading genesis file %s: %v", genFile, err)
		}
		return privKey, genesisCfg, nil
	}

	if !autoGen {
		return nil, nil, fmt.Errorf("genesis file not found: %s", genFile)
	}

	if err = os.MkdirAll(chainRootDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("error creating abci config dir %s: %v", chainRootDir, err)
	}

	genesisCfg = chain.NewGenesisWithValidator(pub)
	err = genesisCfg.SaveAs(genFile)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to write genesis file %s: %v", genFile, err)
	}
	fmt.Printf("Generated new genesis file: %v\n", genFile)
	return privKey, genesisCfg, nil
}

/* TODO: restore when we figure out how to compute appHash with postgres
func setGenesisAppHash(appHash []byte, genesisFile string) error {
	genesisConf, err := LoadGenesisConfig(genesisFile)
	if err != nil {
		return fmt.Errorf("failed to load genesis file: %w", err)
	}

	genesisConf.DataAppHash = appHash
	if err := genesisConf.SaveAs(genesisFile); err != nil {
		return fmt.Errorf("failed to save genesis file: %w", err)
	}
	return nil
}
*/

// PatchGenesisAppHash computes the apphash from a full contents of all sqlite
// files in the provided folder, and if genesis file is provided, updates the
// app_hash in the file.
/* WARNING: this is not complete, only a concept.  We
   don't have SQLite anymore, so the file hashing is replaced with a suggestion
   of what we might do when this is implemented.

func PatchGenesisAppHash(ctx context.Context, cfg *pg.ConnConfig, genesisFile string) ([]byte, error) {
	conns, err := pg.NewPool(ctx, &pg.PoolConfig{
		ConnConfig: *cfg,
		MaxConns:   2,
	})
	if err != nil {
		return nil, err
	}

	engineCtx, err := execution.NewGlobalContext(ctx, conns, map[string]actions.ExtensionInitializer{}, nil)
	if err != nil {
		return nil, err
	}

	datasets, err := engineCtx.ListDatasets(ctx, nil)
	if err != nil {
		return nil, err
	}

	hasher := sha256.New()

	for _, dataset := range datasets {
		schema, err := engineCtx.GetSchema(ctx, dataset.DBID)
		if err != nil {
			return nil, err
		}
		pgSchema := execution.DBIDSchema(dataset.DBID)
		for _, table := range schema.Tables {
			qualifiedTableName := pgSchema + "." + table.Name
			// HERE we need to iterate over all the columns in a deterministic way to form a digest of all the data.
			fmt.Println(qualifiedTableName)
			// ...
		}
	}

	// ALSO, we should probably migrate and digest the accounts table.
	// ...

	// Optionally update the app_hash in the genesis file.
	genesisHash := hasher.Sum(nil)
	if genesisFile != "" {
		err = setGenesisAppHash(genesisHash, genesisFile)
		if err != nil {
			return nil, err
		}
	}

	return genesisHash, nil
}
*/
