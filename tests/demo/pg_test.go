package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// postgresContainer represents the postgres container type used in the module
type postgresContainer struct {
	testcontainers.Container
}

type postgresContainerOption func(req *testcontainers.ContainerRequest)

func WithWaitStrategy(strategies ...wait.Strategy) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(1 * time.Minute)
	}
}

func WithPort(port string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.ExposedPorts = append(req.ExposedPorts, port)
	}
}

func WithInitialDatabase(user string, password string, dbName string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["POSTGRES_USER"] = user
		req.Env["POSTGRES_PASSWORD"] = password
		req.Env["POSTGRES_DB"] = dbName
	}
}

// setupPostgres creates an instance of the postgres container type
func setupPostgres(ctx context.Context, opts ...postgresContainerOption) (*postgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:11-alpine",
		Env:          map[string]string{},
		ExposedPorts: []string{},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
	}

	for _, opt := range opts {
		opt(&req)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &postgresContainer{Container: container}, nil
}

func TestPostgres(t *testing.T) {
	ctx := context.Background()

	const dbname = "test-db"
	const user = "postgres"
	const password = "password"

	port, err := nat.NewPort("tcp", "5432")
	require.NoError(t, err)

	container, err := setupPostgres(ctx,
		WithPort(port.Port()),
		WithInitialDatabase(user, password, dbname),
		WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	containerPort, err := container.MappedPort(ctx, port)
	assert.NoError(t, err)

	host, err := container.Host(ctx)
	assert.NoError(t, err)

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, containerPort.Port(), user, password, dbname)

	// perform assertions
	db, err := sql.Open("pgx", connStr)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	result, err := db.Exec("CREATE TABLE IF NOT EXISTS test (id int, name varchar(255));")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	result, err = db.Exec("INSERT INTO test (id, name) VALUES (1, 'test');")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestContainerWithWaitForSQL(t *testing.T) {
	const dbname = "test-db"
	const user = "postgres"
	const password = "password"

	ctx := context.Background()

	var port = "5432/tcp"
	dbURL := func(host string, port nat.Port) string {
		return fmt.Sprintf("postgres://postgres:password@%s:%s/%s?sslmode=disable", host, port.Port(), dbname)
	}

	t.Run("default query", func(t *testing.T) {
		container, err := setupPostgres(ctx,
			WithPort(port),
			WithInitialDatabase("postgres", "password", dbname),
			WithWaitStrategy(wait.ForSQL(nat.Port(port), "pgx", dbURL)))
		require.NoError(t, err)
		require.NotNil(t, container)
	})
	t.Run("custom query", func(t *testing.T) {
		container, err := setupPostgres(
			ctx,
			WithPort(port),
			WithInitialDatabase(user, password, dbname),
			WithWaitStrategy(wait.ForSQL(nat.Port(port), "pgx", dbURL).WithStartupTimeout(time.Second*5).WithQuery("SELECT 10")),
		)
		require.NoError(t, err)
		require.NotNil(t, container)
	})
	t.Run("custom bad query", func(t *testing.T) {
		container, err := setupPostgres(
			ctx,
			WithPort(port),
			WithInitialDatabase(user, password, dbname),
			WithWaitStrategy(wait.ForSQL(nat.Port(port), "pgx", dbURL).WithStartupTimeout(time.Second*5).WithQuery("SELECT 'a' from b")),
		)
		require.Error(t, err)
		require.Nil(t, container)
	})
}
