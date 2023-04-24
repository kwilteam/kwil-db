package adapters

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"path"
	"runtime"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	PgImage    = "postgres:15-alpine"
	PgPort     = "5432"
	pgUser     = "postgres"
	pgPassword = "postgres"
)

// postgresContainer represents the postgres container type used in the module
type postgresContainer struct {
	TContainer
}

func (c *postgresContainer) GetUnexposedDBUrl(ctx context.Context, databaseName string) string {
	unexposedEndpoint, _ := c.UnexposedEndpoint(ctx)
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable", pgUser, pgPassword, unexposedEndpoint, databaseName)
}

func WithInitialDatabase(user string, password string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env["POSTGRES_USER"] = user
		req.Env["POSTGRES_PASSWORD"] = password
		// req.Env["POSTGRES_DB"] = dbName
	}
}

// setupPostgres creates an instance of the postgres container type
func setupPostgres(ctx context.Context, opts ...containerOption) (*postgresContainer, error) {
	_, currentFilePath, _, _ := runtime.Caller(1)

	req := testcontainers.ContainerRequest{
		Name:  fmt.Sprintf("postgres-%d", time.Now().Unix()),
		Image: PgImage,
		Env:   map[string]string{},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      path.Join(path.Dir(currentFilePath), "../../../scripts/pg-init-scripts/initdb.sh"),
				ContainerFilePath: "/docker-entrypoint-initdb.d/initdb.sh",
				FileMode:          0644,
			},
		},
		Networks:     []string{"test-network"},
		ExposedPorts: []string{}}

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

	return &postgresContainer{TContainer{
		Container:     container,
		ContainerPort: PgPort,
	}}, nil
}

func StartDBDockerService(t *testing.T, ctx context.Context) *postgresContainer {
	//t.Helper()

	container, err := setupPostgres(ctx,
		WithNetwork(kwilTestNetworkName),
		WithExposedPort(PgPort),
		WithInitialDatabase(pgUser, pgPassword),
		WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(time.Second*20)))

	require.NoError(t, err, "Could not start postgres container")

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx), "Could not stop postgres container")
	})

	err = container.ShowPortInfo(ctx)
	require.NoError(t, err)

	return container
}
