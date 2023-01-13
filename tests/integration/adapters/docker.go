package adapters

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func StartDBDockerService(t *testing.T, ctx context.Context) *postgresContainer {
	t.Helper()

	//dbURL := func(host string, port nat.Port) string {
	//	return fmt.Sprintf("postgres://%:%s@%s:%s/%s?sslmode=disable", pgUser, pgPassword, host, port.Port(), kwildDatabase)
	//}

	container, err := setupPostgres(ctx,
		WithNetwork(kwildTestNetworkName),
		WithPort(PgPort),
		WithInitialDatabase(pgUser, pgPassword),
		WithFiles(map[string]string{"../../../../scripts/pg-init-scripts/initdb.sh": "/docker-entrypoint-initdb.d/initdb.sh"}),
		WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(time.Second*20)))
	//wait.ForSQL(nat.Port(PgPort), "pgx", dbURL).WithStartupTimeout(time.Second*30)))

	assert.NoError(t, err, "Could not start postgres container")

	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx), "Could not stop postgres container")
	})

	err = container.ShowInfo(ctx)
	assert.NoError(t, err)

	//connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	//	host, portC.Port(), pgUser, pgPassword, kwildDatabase)
	//
	//// perform assertions
	//db, err := sql.Open("pgx", connStr)
	//assert.NoError(t, err)
	//assert.NotNil(t, db)
	//defer db.Close()
	//
	//_, err = db.Exec("insert into accounts (account_address, balance) values ($1, $2)", "0x123", 100)
	//assert.NoError(t, err)
	//
	//r, err := db.Query("select account_address from accounts")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//defer r.Close()
	//for r.Next() {
	//	var accountAddress string
	//	err = r.Scan(&accountAddress)
	//	if err != nil {
	//		break
	//	}
	//	t.Log("???????", accountAddress)
	//}

	return container
}

func StartKwildDockerService(t *testing.T, ctx context.Context, dbContainer *postgresContainer) *kwildContainer {
	t.Helper()
	ipC, _ := dbContainer.ContainerIP(ctx)

	container, err := setupKwild(ctx,
		//WithDockerFile("kwild"),
		WithNetwork(kwildTestNetworkName),
		WithPort(KwildPort),
		WithEnv(map[string]string{
			"PG_DATABASE_URL": fmt.Sprintf(
				"postgres://%s:%s@%s:%s/%s?sslmode=disable", pgUser, pgPassword, ipC, PgPort, kwildDatabase),
		}),
		// ForListeningPort requires image has /bin/sh
		WithWaitStrategy(wait.ForLog("grpc server started")))

	assert.NoError(t, err, "Could not start kwild container")

	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx), "Could not stop kwild container")
	})

	err = container.ShowInfo(ctx)
	assert.NoError(t, err)
	return container
}

func StartDockerServer(t *testing.T, port string, cmd string) testcontainers.Container {
	t.Helper()

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: newTCDockerfile(cmd),
		ExposedPorts:   []string{fmt.Sprintf("%s:%s", port, port)},
		WaitingFor:     wait.ForListeningPort(nat.Port(port)).WithStartupTimeout(startupTimeout),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	assert.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, container.Terminate(ctx))
	})

	return container
}

func newTCDockerfile(cmd string) testcontainers.FromDockerfile {
	return testcontainers.FromDockerfile{
		Context:    "../../.",
		Dockerfile: fmt.Sprintf("docker/%s.dockerfile", cmd),
		BuildArgs: map[string]*string{
			"bin_to_build": &cmd,
		},
		PrintBuildLog: true,
	}
}
