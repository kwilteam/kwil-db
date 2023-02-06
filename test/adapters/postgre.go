package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	PgPort     = "5432"
	pgUser     = "postgres"
	pgPassword = "postgres"
)

// postgresContainer represents the postgres container type used in the module
type postgresContainer struct {
	TContainer
}

func WithInitialDatabase(user string, password string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env["POSTGRES_USER"] = user
		req.Env["POSTGRES_PASSWORD"] = password
		//req.Env["POSTGRES_DB"] = dbName
	}
}

// setupPostgres creates an instance of the postgres container type
func setupPostgres(ctx context.Context, opts ...containerOption) (*postgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Name:         fmt.Sprintf("postgres-%d", time.Now().Unix()),
		Image:        "postgres:15-alpine",
		Env:          map[string]string{},
		Files:        []testcontainers.ContainerFile{},
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

func (c *postgresContainer) test(ctx context.Context) error {
	host, err := c.Host(ctx)
	if err != nil {
		return err
	}
	portC, err := c.MappedPort(context.Background(), nat.Port(c.ContainerPort))
	if err != nil {
		return err
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, portC.Port(), pgUser, pgPassword, kwildDatabase)

	// perform assertions
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return err
	}
	if db == nil {
		return err
	}
	defer db.Close()

	r, err := db.Query("select account_address from accounts")
	if err != nil {
		return err
	}
	defer r.Close()
	for r.Next() {
		var accountAddress string
		err = r.Scan(&accountAddress)
		if err != nil {
			break
		}
		fmt.Println("got config ", accountAddress)
	}

	return nil
}
