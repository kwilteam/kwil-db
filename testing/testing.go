// package testing provides tools for testing Kuneiform schemas.
// It is meant to be used by consumers of Kwil to easily test schemas
// in a fully synchronous environment.
package testing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cometbft/cometbft/test/e2e/pkg/exec"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/stretchr/testify/assert"
)

// RunSchemaTest runs a SchemaTest.
// It is meant to be used with Go's testing package.
func RunSchemaTest(t *testing.T, s SchemaTest) {
	err := s.Run(context.Background(), &Options{
		UseTestContainer: true,
		Logger:           t,
	})
	if err != nil {
		t.Fatalf("test failed: %s", err.Error())
	}
}

// SchemaTest allows for testing schemas against a live database.
// It allows for several ways of specifying schemas to deploy, as well
// as functions that can be run against the schemas, and expected results.
type SchemaTest struct {
	// Name is the name of the test case.
	Name string `json:"name"`
	// Schemas are plain text schemas to deploy as
	// part of the text.
	Schemas []string `json:"-"`
	// SchemaFiles are paths to the schema files to deploy.
	SchemaFiles []string `json:"schema_files"`
	// SeedStatements are SQL statements run before each test that are
	// meant to seed the database with data. It maps the database name
	// to the SQL statements to run. The name is the database name,
	// defined using "database <name>;". The test case will derive the
	// DBID from the name.
	SeedStatements map[string][]string `json:"seed_statements"`
	// TestCases execute actions or procedures against the database
	// engine, taking certain inputs and expecting certain outputs or
	// errors. These run separately from the functions, and separately
	// from each other. They are the easiest way to test the database
	// engine, but if more nuanced tests are needed (e.g. to simulate
	// several different wallets), the FunctionTests field should be used
	// instead. All schemas will be redeployed and all seed data re-applied
	// between executing each TestCase.
	TestCases []TestCase `json:"test_cases"`
	// FunctionTests are arbitrary functions that can be used to
	// execute any logic against the schemas.
	// All schemas will be reset before each function is run.
	// FunctionTests are more cumbersome to use than TestCases, but
	// they allow for more nuanced testing and flexibility.
	// All functions and testcases are run against fresh schemas.
	FunctionTests []TestFunc `json:"-"`
}

// Run runs the test case.
// If opts is nil, the test set up and teardown create a Docker
// testcontainer to run the test.
func (tc SchemaTest) Run(ctx context.Context, opts *Options) error {
	if opts == nil {
		opts = &Options{}

		// doing this here since doing it outside
		// of the nil check would make it impossible to tell if
		// there was a user config error, or if we just need defaults.
		opts.UseTestContainer = true
	}

	if opts.Logger == nil {
		l := log.New(log.Config{
			Level: "info",
		})
		opts.Logger = &kwilLoggerWrapper{
			Logger: &l,
		}
	}

	err := opts.valid()
	if err != nil {
		return fmt.Errorf("test configuration error: %w", err)
	}

	schemas := tc.Schemas
	for _, schemaFile := range tc.SchemaFiles {
		bts, err := os.ReadFile(schemaFile)
		if err != nil {
			return err
		}
		schemas = append(schemas, string(bts))
	}

	var parsedSchemas []*types.Schema
	for _, schema := range schemas {
		s, err := parse.Parse([]byte(schema))
		if err != nil {
			return fmt.Errorf(`error parsing schema: %w`, err)
		}
		parsedSchemas = append(parsedSchemas, s)

		s.Owner = deployer

		opts.Logger.Logf(`using schema "%s" (DBID: "%s")`, s.Name, s.DBID())
	}

	// connect to Postgres, and run each test case in its
	// own transaction that is rolled back.
	return runWithPostgres(ctx, opts, func(ctx context.Context, d *pg.DB, logger Logger) error {
		testFns := tc.FunctionTests
		var testFnIdentifiers []string // tracks an identifier for each sub test
		var testNames []string         // tracks the names of each sub test

		// identify the functions
		for i := range tc.FunctionTests {
			testFnIdentifiers = append(testFnIdentifiers, fmt.Sprintf("TestCase.Function-%d", i))
		}

		// identify the executions
		for _, tc := range tc.TestCases {
			tc2 := tc // copy to avoid loop variable capture
			testFns = append(testFns, tc2.runExecution)
			testFnIdentifiers = append(testFnIdentifiers, fmt.Sprintf("TestCase.Execution: %s", tc2.Name))
			testNames = append(testNames, tc2.Name)
		}

		var errs []error

		for i, testFn := range testFns {
			// each test case is named after the index it is for its type.
			// It is run in a function to allow defers
			err := func() error {
				logger.Logf(`running test "%s"`, testFnIdentifiers[i])

				// setup a tx and execution engine
				outerTx, err := d.BeginOuterTx(ctx)
				if err != nil {
					return err
				}
				// always rollback the outer transaction to reset the database
				defer outerTx.Rollback(ctx)

				err = execution.InitializeEngine(ctx, outerTx)
				if err != nil {
					return err
				}

				var logger log.SugaredLogger
				// if this is a kwil logger, we can keep using it.
				// If it is from testing.T, we should make a Kwil logger.
				if wrapped, ok := opts.Logger.(*kwilLoggerWrapper); ok {
					logger = wrapped.Sugar()
				} else {
					logger = log.New(log.Config{
						Level: "info",
					}).Sugar()
				}

				engine, err := execution.NewGlobalContext(ctx, outerTx, maps.Clone(precompiles.RegisteredPrecompiles()), &common.Service{
					Logger:           logger,
					ExtensionConfigs: map[string]map[string]string{},
					Identity:         []byte("node"),
				})
				if err != nil {
					return err
				}

				platform := &Platform{
					Engine:   engine,
					DB:       outerTx,
					Deployer: deployer,
					Logger:   opts.Logger,
				}

				// deploy schemas
				for _, schema := range parsedSchemas {
					err := engine.CreateDataset(ctx, outerTx, schema, &common.TransactionData{
						Signer: deployer,
						Caller: string(deployer),
						TxID:   platform.Txid(),
						Height: 0,
					})
					if err != nil {
						return err
					}
				}

				// seed data
				for dbName, seed := range tc.SeedStatements {
					if strings.HasSuffix(dbName, ".kf") {
						// while I was testing this, I hit this twice by accident, so I
						// figured I should add in a helpful error message
						return fmt.Errorf(`seed statement target must be the schema name, not the file name. Received "%s"`, dbName)
					}

					for _, sql := range seed {
						dbid := utils.GenerateDBID(dbName, deployer)
						_, err = engine.Execute(ctx, outerTx, dbid, sql, nil)
						if err != nil {
							return fmt.Errorf(`error executing seed query "%s" on schema "%s": %s`, sql, dbName, err)
						}
					}
				}

				// run test function
				err = testFn(ctx, platform)
				if err != nil {
					return fmt.Errorf(`test "%s" failed: %w`, testNames[i], err)
				}
				return nil
			}()
			if err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) == 0 {
			return nil
		}
		return errors.Join(errs...)
	})
}

var deployer = []byte("deployer")

// TestFunc is a function that can be run against the database engine.
// A returned error signals a failed test.
type TestFunc func(ctx context.Context, platform *Platform) error

// TestCase executes an action or procedure against the database engine.
// It can be given inputs, expected outputs, expected error types,
// and expected error messages.
type TestCase struct {
	// Name is a name that the test will be identified by if it fails.
	Name string `json:"name"`
	// Database is the name of the database schema to execute the
	// action/procedure against. This is the database NAME,
	// defined using "database <name>;". The test case will
	// derive the DBID from the name.
	Database string `json:"database"`
	// Name is the name of the action/procedure.
	Target string `json:"target"`
	// Args are the inputs to the action/procedure.
	// If the action/procedure takes no parameters, this should be nil.
	Args []any `json:"args"`
	// Returns are the expected outputs of the action/procedure.
	// It takes a two-dimensional array to model the output of a table.
	// If the action/procedure has no outputs, this should be nil.
	Returns [][]any `json:"returns"`
	// Err is the expected error type. If no error is expected, this
	// should be nil.
	Err error `json:"-"`
	// ErrMsg will search the error returned by the action/procedure for
	// the given substring. If no error is expected, this should be an
	// empty string.
	ErrMsg string `json:"error"`
	// Signer sets the @caller, and the bytes will be used as the @signer.
	// If empty, the test case schema deployer will be used.
	Caller string `json:"caller"`
	// BlockHeight sets the blockheight for the test, accessible by
	// the @height variable. If not set, it will default to 0.
	Height int64 `json:"height"`
}

// run runs the Execution as a TestFunc
func (e *TestCase) runExecution(ctx context.Context, platform *Platform) error {
	caller := string(deployer)
	if e.Caller != "" {
		caller = e.Caller
	}

	dbid := utils.GenerateDBID(e.Database, deployer)

	// log to help users debug failed tests
	platform.Logger.Logf(`executing action/procedure "%s" against schema "%s" (DBID: "%s")`, e.Target, e.Database, dbid)

	res, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
		TransactionData: common.TransactionData{
			Signer: []byte(caller),
			Caller: caller,
			Height: e.Height,
			TxID:   platform.Txid(),
		},
		Dataset:   dbid,
		Procedure: e.Target,
		Args:      e.Args,
	})
	if err != nil {
		// if error is not nil, the test should only pass if either
		// Err or ErrMsg or both is set
		expectsErr := false
		if e.Err != nil {
			expectsErr = true
			errTypeName := reflect.TypeOf(e.Err).Elem().Name()
			if !errors.Is(err, e.Err) {
				return fmt.Errorf(`expected error of type "%s", received error: %w`, errTypeName, err)
			}
		}
		if e.ErrMsg != "" {
			expectsErr = true
			if !strings.Contains(err.Error(), e.ErrMsg) {
				return fmt.Errorf(`expected error message to contain substring "%s", received error: %w`, e.ErrMsg, err)
			}
		}

		if !expectsErr {
			return fmt.Errorf(`unexpected error: %w`, err)
		}

		return nil
	}

	if len(res.Rows) != len(e.Returns) {
		return fmt.Errorf("expected %d rows to be returned, received %d", len(e.Returns), len(res.Rows))
	}

	for i, row := range res.Rows {
		if len(row) != len(e.Returns[i]) {
			return fmt.Errorf("expected %d columns to be returned, received %d", len(e.Returns[i]), len(row))
		}

		for j, col := range row {
			if !assert.ObjectsAreEqualValues(e.Returns[i][j], col) {
				return fmt.Errorf("incorrect value for expected result: row %d, column %d", i, j)
			}
		}
	}

	return nil
}

// Platform provides utilities and info for usage in test functions.
// It allows users to access the database engine, get information about the
// schema deployers, control transactions, or even directly access PostgreSQL.
type Platform struct {
	// Engine is the Kuneiform engine that can deploy schemas, execute actions/procedures,
	// execute adhoc SQL, and more. It should be the primary way to interact with the database.
	Engine common.Engine
	// DB is the database engine that the test case is running against.
	// It provides access directly to Postgres, and has superuser access
	// to the underlying database. If users want to execute ad-hoc queries,
	// they should prefer to use the Engine, which parses Kwil's SQL standard,
	// and guarantees determinism.
	DB sql.DB
	// Deployer is the public identifier of the user that deployed the schemas
	// during test setup. It can be used to execute owner-only actions and procedures.
	// To execute owner-only actions and procedures, set the Deployer to be the
	// *common.ExecutionData.TransactionData.Signer field when executing the
	// action/procedure.
	Deployer []byte

	// Logger is for logging information during execution of the test.
	Logger Logger

	// lastTxid is the last transaction ID that was used.
	lastTxid []byte
}

// Txid returns a new, unused transaction ID.
// It is deterministic, making tests repeatable.
func (p *Platform) Txid() string {
	if len(p.lastTxid) == 0 {
		b := sha256.Sum256([]byte("first txid"))
		p.lastTxid = b[:]
		return hex.EncodeToString(b[:])
	}

	b := sha256.Sum256(p.lastTxid)
	p.lastTxid = b[:]
	return hex.EncodeToString(b[:])
}

// runWithPostgres runs the callback function with a postgres container.
func runWithPostgres(ctx context.Context, opts *Options, fn func(context.Context, *pg.DB, Logger) error) (err error) {
	if !opts.UseTestContainer {
		db, err := pg.NewDB(ctx, &pg.DBConfig{
			PoolConfig: pg.PoolConfig{
				MaxConns: 11,
				ConnConfig: pg.ConnConfig{
					Host:   opts.Conn.Host,
					Port:   opts.Conn.Port,
					User:   opts.Conn.User,
					Pass:   opts.Conn.Pass,
					DBName: opts.Conn.DBName,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("error setting up database: %w", err)
		}

		defer db.Close()

		return fn(ctx, db, opts.Logger)
	}

	// check if the user has docker
	err = exec.Command(ctx, "docker")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("docker not found. Please ensure Docker is installed and running")
		}
		return fmt.Errorf("error checking for Docker installation: %w", err)
	}

	port := "52853"

	// Run the container
	bts, err := exec.CommandOutput(ctx, "docker", "run", "-d", "-p", fmt.Sprintf("%s:5432", port), "--name", "kwil-testing-postgres", "-e",
		"POSTGRES_HOST_AUTH_METHOD=trust", "kwildb/postgres:latest")
	if err != nil {
		return fmt.Errorf("error running test container: %w", err)
	}
	defer func() {
		err2 := exec.Command(ctx, "docker", "rm", "-f", "kwil-testing-postgres")
		if err2 != nil {
			if err == nil {
				err = err2
			} else {
				err = errors.Join(err, err2)
			}
		}
	}()

	opts.Logger.Logf("running test container: %s", string(bts))

	time.Sleep(1 * time.Second) // stupid hack needed for the container to be ready

	db, err := pg.NewDB(ctx, &pg.DBConfig{
		PoolConfig: pg.PoolConfig{
			MaxConns: 11,
			ConnConfig: pg.ConnConfig{
				Host:   "localhost",
				Port:   port,
				User:   "kwild",
				Pass:   "kwild", // would be ignored if pg_hba.conf set with trust
				DBName: "kwil_test_db",
			},
		},
	})
	if err != nil {
		return err
	}

	defer db.Close()

	return fn(ctx, db, opts.Logger)
}

// Options configures optional parameters for running the test.
// Either UseTestContainer should be true, or a valid
// PostgreSQL connection should be specified.
type Options struct {
	// UseTestContainer specifies whether the test should setup and
	// teardown a test container.
	UseTestContainer bool
	// Conn specifies a manually setup Postgres connection that the
	// test can connect to.
	Conn *pg.ConnConfig
	// Logger is a logger to be used in the test
	Logger Logger
}

func (d *Options) valid() error {
	if d.UseTestContainer && d.Conn != nil {
		return fmt.Errorf("test cannot both use a test container and specify a Postgres connection")
	}

	if !d.UseTestContainer && d.Conn == nil {
		return fmt.Errorf("test must either use a test container or specify a Postgres connection")
	}

	return nil
}

// Logger is a logger that the tests use while running.
// It can be made to fit both Kwil's Logger interface,
// as well as Go's stdlib test package
type Logger interface {
	Logf(string, ...any)
}

// LoggerFromKwilLogger wraps the Kwil standard logger so
// so that it can be used in tests
func LoggerFromKwilLogger(log *log.Logger) Logger {
	return &kwilLoggerWrapper{
		Logger: log,
	}
}

type kwilLoggerWrapper struct {
	*log.Logger
}

func (k *kwilLoggerWrapper) Logf(s string, a ...any) {
	k.Logger.Logf(k.L.Level(), s, a...)
}
