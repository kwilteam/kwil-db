package snapshot

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/internal/abci/meta"
	"github.com/kwilteam/kwil-db/internal/sql/pg"

	"github.com/spf13/cobra"
)

var (
	createLongExplain = `
This command creates a logical snapshot and prepares a genesis.json for a new network based on that snapshot. This command interacts directly with the underlying PostgreSQL server, bypassing any interactions with the kwild node. It requires a database user with superuser privileges.

The snapshots generated by kwild during normal operation are different from the snapshots generated by this tool. This tool is to prepare a new network based on the final state of an existing network, while kwild's snapshots support fast "state sync" for nodes joining an existing network.`

	createExample = `# Create database snapshot and the genesis file to initialize a new network
# Password is optional if the db is operating in a trust authentication mode.
kwil-admin snapshot create --dbname kwildb --user user1 --password pass1 --host localhost --port 5432 --snapdir /path/to/snapshot/dir

# Snapshot and genesis files will be created in the snapshot directory
ls /path/to/snapshot/dir
genesis.json    kwildb-snapshot.sql.gz`
)

/*
Use this at the beginning of the sql dump file to drop any active connections on the 'kwild' database.
Not a good idea to use this if the Kwild node is connected to this database.

SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = 'kwild'
  AND pid <> pg_backend_pid();
*/

func createCmd() *cobra.Command {
	var snapshotDir, dbName, dbUser, dbPass, dbHost, dbPort string
	var maxRowSize int
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Creates a snapshot of the database.",
		Long:    createLongExplain,
		Example: createExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshotDir, err := common.ExpandPath(snapshotDir)
			if err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to expand snapshot directory path: %v", err))
			}

			height, logs, err := pgDump(cmd.Context(), dbName, dbUser, dbPass, dbHost, dbPort, snapshotDir)
			if err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to create database snapshot: %v", err))
			}

			r := &createSnapshotRes{Logs: logs, Height: height}
			return display.PrintCmd(cmd, r)
		},
	}

	cmd.Flags().StringVar(&snapshotDir, "snapdir", "kwild-snaps", "Directory to store the snapshot and hash files")
	cmd.Flags().StringVar(&dbName, "dbname", "kwild", "Name of the database to snapshot")
	cmd.Flags().StringVar(&dbUser, "user", "postgres", "User with administrative privileges on the database")
	cmd.Flags().StringVar(&dbPass, "password", "", "Password for the database user")
	cmd.Flags().StringVar(&dbHost, "host", "localhost", "Host of the database")
	cmd.Flags().StringVar(&dbPort, "port", "5432", "Port of the database")

	// TODO: Deprecate below flags
	cmd.Flags().IntVar(&maxRowSize, "max-row-size", 4*1024*1024, "Maximum row size to read from pg_dump (default: 4MB). Adjust this accordingly if you encounter 'bufio.Scanner: token too long' error.")
	cmd.Flags().MarkDeprecated("max-row-size", "max-row-size has no more influence on the snapshot creation process. It is deprecated and will be removed in v0.10.0")
	return cmd
}

type createSnapshotRes struct {
	Logs   []string `json:"logs"`
	Height int64    `json:"height"`
}

func (c *createSnapshotRes) MarshalJSON() ([]byte, error) {
	return json.Marshal(c)
}

func (c *createSnapshotRes) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("Snapshot created successfully at height: %d \n%s", c.Height, strings.Join(c.Logs, "\n"))), nil
}

// PGDump uses pg_dump to create a snapshot of the database.
// It returns messages to log and an error if any.
func pgDump(ctx context.Context, dbName, dbUser, dbPass, dbHost, dbPort string, snapshotDir string) (height int64, logs []string, err error) {
	// Get the chain height
	height, err = chainHeight(ctx, dbName, dbUser, dbPass, dbHost, dbPort)
	if err != nil {
		return -1, nil, fmt.Errorf("failed to get chain height: %w", err)
	}
	// Check if the snapshot directory exists, if not create it
	err = os.MkdirAll(snapshotDir, 0755)
	if err != nil {
		return -1, nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	dumpFile := filepath.Join(snapshotDir, "kwildb-snapshot.sql.gz")
	outputFile, err := os.Create(dumpFile)
	if err != nil {
		return -1, nil, fmt.Errorf("failed to create dump file: %w", err)
	}
	// delete the dump file if an error occurs anywhere during the snapshot process
	defer func() {
		outputFile.Close()
		if err != nil {
			os.Remove(dumpFile)
		}
	}()

	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	pgDumpCmd := exec.CommandContext(ctx,
		"pg_dump",
		"--dbname", dbName,
		"--format", "plain",
		"--schema", "kwild_voting", // Include only the processed table
		// Voting Schema (remove the COPY command from voters table during sanitization)

		// Account Schema
		"--schema", "kwild_accts",

		// Internal Schema
		"--schema", "kwild_internal",
		"-T", "kwild_internal.sentry", // Exclude sentry table (no versioning)

		// User Schemas
		"--schema", "ds_*",

		// kwild_chain is not included in this snapshot, as this is used for genesis state

		// Other options
		"--no-unlogged-table-data",
		"--no-comments",
		"--create",
		// "--clean", // drops database first before adding commands to create it
		// "--if-exists",
		"--no-publications",
		"--no-unlogged-table-data",
		"--no-tablespaces",
		"--no-table-access-method",
		"--no-security-labels",
		"--no-subscriptions",
		"--large-objects",
		"--no-owner", // Do not include ownership information, to restore on any user
		"--username", dbUser,
		"--host", dbHost,
		"--port", dbPort,
		"--no-password",
	)
	pgDumpCmd.Env = append(os.Environ(), "PGPASSWORD="+dbPass)

	var stderr bytes.Buffer
	pgDumpOutput, err := pgDumpCmd.StdoutPipe()
	if err != nil {
		return -1, nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	pgDumpCmd.Stderr = &stderr

	if err := pgDumpCmd.Start(); err != nil {
		return -1, nil, fmt.Errorf("failed to start pg_dump command: %w", err)
	}
	defer pgDumpOutput.Close()

	hasher := sha256.New()
	var inVotersBlock, schemaStarted bool
	var validatorCount int64
	genCfg := chain.DefaultGenesisConfig()
	genCfg.Alloc = make(map[string]*big.Int)
	multiWriter := io.MultiWriter(gzipWriter, hasher)
	var totalBytes int64

	// Sanitize the output of pg_dump to include only the necessary tables and data
	reader := bufio.NewReader(pgDumpOutput)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return -1, nil, fmt.Errorf("failed to read pg_dump output: %w", err)
		}

		trimLine := strings.TrimSpace(line)

		// Remove whitespaces, set and select statements, process voters table
		if inVotersBlock {
			// Example voter: \\xdae5e91f74b95a9db05fc0f1f8c07f95	\\x9e52ff636caf4988e72e4ac865e6ef83a1e262d1a6376a300f3db8884e1f2253	1

			if trimLine == "\\." { // End of voters block
				inVotersBlock = false
				n, err := multiWriter.Write([]byte(line)) // line includes the \n
				if err != nil {
					return -1, nil, fmt.Errorf("failed to write to gzip writer: %w", err)
				}
				totalBytes += int64(n)
				continue
			}

			strs := strings.Split(trimLine, "\t")
			if len(strs) != 3 {
				return -1, nil, fmt.Errorf("invalid voter line: %s", trimLine)
			}
			voterID, err := hex.DecodeString(strs[1][3:]) // Remove the leading \\x
			if err != nil {
				return -1, nil, fmt.Errorf("failed to decode voter ID: %w", err)
			}

			power, err := strconv.ParseInt(strs[2], 10, 64)
			if err != nil {
				return -1, nil, fmt.Errorf("failed to parse power: %w", err)
			}

			genCfg.Validators = append(genCfg.Validators, &chain.GenesisValidator{
				PubKey: voterID,
				Power:  power,
				Name:   fmt.Sprintf("validator-%d", validatorCount),
			})
			validatorCount++
		} else {
			if line == "" || trimLine == "" { // Skip empty lines
				continue
			} else if strings.HasPrefix(trimLine, "--") { // Skip comments
				continue
			} else if !schemaStarted && (strings.HasPrefix(trimLine, "SET") || strings.HasPrefix(trimLine, "SELECT") || strings.HasPrefix(trimLine, "\\connect") || strings.HasPrefix(trimLine, "CREATE DATABASE")) {
				// Skip SET and SELECT and connect and create database statements
				continue
			} else {
				// Start of schema
				if !schemaStarted && (strings.HasPrefix(trimLine, "CREATE SCHEMA") || strings.HasPrefix(trimLine, "CREATE TABLE") || strings.HasPrefix(trimLine, "CREATE FUNCTION")) {
					schemaStarted = true
				}

				if strings.HasPrefix(trimLine, "COPY kwild_voting.voters") && strings.Contains(trimLine, "FROM stdin;") {
					inVotersBlock = true
				}

				// Write the sanitized line to the gzip writer
				n, err := multiWriter.Write([]byte(line)) // line includes the \n
				if err != nil {
					return -1, nil, fmt.Errorf("failed to write to gzip writer: %w", err)
				}
				totalBytes += int64(n)
			}
		}
	}

	// Close the writer when pg_dump completes to signal EOF to sed
	if err := pgDumpCmd.Wait(); err != nil {
		return -1, nil, errors.New(stderr.String())
	}

	// Append the below sql statement to the dump file to adjust the expiration times of the resolutions
	// This is to ensure that the resolutions are correctly expired on the new network
	n, err := multiWriter.Write([]byte("UPDATE kwild_voting.resolutions SET expiration = expiration-" + strconv.FormatInt(height, 10) + ";\n"))
	if err != nil {
		return -1, nil, fmt.Errorf("failed to write resolution updates to gzip writer: %w", err)
	}

	totalBytes += int64(n)

	gzipWriter.Flush()
	hash := hasher.Sum(nil)
	genCfg.DataAppHash = hash

	// Write the genesis config to a file
	genesisFile := filepath.Join(snapshotDir, "genesis.json")
	if err := genCfg.SaveAs(genesisFile); err != nil {
		return -1, nil, fmt.Errorf("failed to save genesis config: %w", err)
	}

	return height, []string{fmt.Sprintf("Snapshot created at: %s, Total bytes written: %d", dumpFile, totalBytes),
		fmt.Sprintf("Genesis config created at: %s, Genesis hash: %s", genesisFile, fmt.Sprintf("%x", hash))}, nil
}

func chainHeight(ctx context.Context, dbName, dbUser, dbPass, dbHost, dbPort string) (int64, error) {
	cfg := &pg.PoolConfig{
		ConnConfig: pg.ConnConfig{
			Host:   dbHost,
			Port:   dbPort,
			User:   dbUser,
			Pass:   dbPass,
			DBName: dbName,
		},
		MaxConns: 2,
	}
	pool, err := pg.NewPool(ctx, cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to create pool: %w", err)
	}
	defer pool.Close()

	height, _, err := meta.GetChainState(ctx, pool)
	if err != nil {
		return 0, fmt.Errorf("failed to get chain state: %w", err)
	}

	return height, nil
}
