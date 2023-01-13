package adapters

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
)

const (
	KwildPort     = "50051"
	kwildDatabase = "kwil"
	kwildImage    = "kwild:latest"
)

// kwildContainer represents the kwild container type used in the module
type kwildContainer struct {
	testcontainers.Container
	Port string
}

// setupKwild creates an instance of the kwild container type
func setupKwild(ctx context.Context, opts ...containerOption) (*kwildContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        kwildImage,
		Env:          map[string]string{},
		Files:        []testcontainers.ContainerFile{},
		ExposedPorts: []string{},
		//Cmd:          []string{"-h"},
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

	return &kwildContainer{Container: container, Port: KwildPort}, nil
}

func (c *kwildContainer) ShowInfo(ctx context.Context) error {
	ipC, err := c.ContainerIP(ctx)
	if err != nil {
		return err
	}
	host, err := c.Host(ctx)
	if err != nil {
		return err
	}
	portC, err := c.MappedPort(context.Background(), KwildPort)
	if err != nil {
		return err
	}
	fmt.Printf("kwild container started, serve at %s:%s, exposed at %s:%s", ipC, KwildPort, host, portC.Port())
	return nil
}
