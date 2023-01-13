package adapters

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

const (
	PgPort     = "5432"
	pgUser     = "postgres"
	pgPassword = "postgres"
)

// postgresContainer represents the postgres container type used in the module
type postgresContainer struct {
	testcontainers.Container
	Port string
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

//// setupPostgres creates an instance of the postgres container type
//func setupPostgres(ctx context.Context, opts ...containerOption) (*postgresContainer, error) {
//	req := testcontainers.ContainerRequest{
//		Image:        "postgres:11-alpine",
//		Env:          map[string]string{},
//		ExposedPorts: []string{},
//		Cmd:          []string{"postgres", "-c", "fsync=off"},
//	}
//
//	for _, opt := range opts {
//		opt(&req)
//	}
//
//	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
//		ContainerRequest: req,
//		Started:          true,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	return &postgresContainer{Container: container}, nil
//}

// setupPostgres creates an instance of the postgres container type
func setupPostgres(ctx context.Context, opts ...containerOption) (*postgresContainer, error) {
	req := testcontainers.ContainerRequest{
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

	return &postgresContainer{Container: container, Port: PgPort}, nil
}

func (c *postgresContainer) ShowInfo(ctx context.Context) error {
	ipC, err := c.ContainerIP(ctx)
	if err != nil {
		return err
	}
	host, err := c.Host(ctx)
	if err != nil {
		return err
	}
	portC, err := c.MappedPort(context.Background(), nat.Port(c.Port))
	if err != nil {
		return err
	}
	fmt.Printf("kwild container started, serve at %s:%s, exposed at %s:%s", ipC, c.Port, host, portC.Port())
	return nil
}
