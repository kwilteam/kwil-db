package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/cometbft/cometbft/types"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/spf13/cobra"
)

var genesisHashCmd = &cobra.Command{
	Use:   "genesis-hash [db-dir]",
	Short: "Generates genesis hash of the sqlite db and if genesis file provided, updates the app_hash in the genesis file",
	Long: `Generates genesis hash of the sqlite db and if genesis file provided, updates the app_hash in the genesis file.
If genesis file is not provided, only the genesis hash is printed to stdout.`,
	Args: cobra.ExactArgs(1),
	RunE: genesisHash,
}

var genesisFile string

func NewGenesisHashCmd() *cobra.Command {
	testnetCmd.Flags().StringVar(&genesisFile, "genesis-file", "", "genesis file to update the app_hash in")
	return genesisHashCmd
}

func genesisHash(cmd *cobra.Command, args []string) error {
	dbDir := args[0]
	// List all sqlite files in the given dir in lexicographical order
	files, err := listFilesAlphabetically(filepath.Join(dbDir, "*.sqlite"))
	if err != nil {
		return err
	}

	// Generate DB Hash
	var cumHash []byte
	for _, file := range files {
		hash, err := fileHash(file)
		if err != nil {
			return err
		}
		cumHash = append(cumHash, hash...)
	}
	genesisHash := crypto.Sha256(cumHash)
	fmt.Println("Genesis Hash: ", hex.EncodeToString(genesisHash))

	// If genesis file provided, update the app_hash in the genesis file
	if genesisFile != "" {
		genesisDoc, err := types.GenesisDocFromFile(genesisFile)
		if err != nil {
			return err
		}
		genesisDoc.AppHash = genesisHash
		if err := genesisDoc.SaveAs(genesisFile); err != nil {
			return err
		}
	}

	return nil
}

func listFilesAlphabetically(filePath string) ([]string, error) {
	files, err := filepath.Glob(filePath)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func fileHash(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash := sha256.New()
	// Reads the file in chunks of 32kb - internally
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}
