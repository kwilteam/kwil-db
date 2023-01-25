package adapters

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

//func init() {
//	os.Setenv("DOCKER_HOST", fmt.Sprintf("unix://%s/.docker/run/docker.sock", os.Getenv("HOME")))
//}

const (
	kwilTestNetworkName = "kwil-test-network"
	startupTimeout      = 1 * time.Minute
)

type containerOption func(req *testcontainers.ContainerRequest)

// TContainer represents the test container type used in the module
type TContainer struct {
	testcontainers.Container
	ContainerPort string
}

func (c *TContainer) ShowPortInfo(ctx context.Context) error {
	exposedEndpoint, err := c.ExposedEndpoint(ctx)
	if err != nil {
		return err
	}

	unexposedEndpoint, err := c.UnexposedEndpoint(ctx)
	if err != nil {
		return err
	}

	name, err := c.Name(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("container %s started, serve at %s, exposed at %s\ns", name, unexposedEndpoint, exposedEndpoint)
	return nil
}

func (c *TContainer) ExposedEndpoint(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	hostPort, err := c.MappedPort(context.Background(), nat.Port(c.ContainerPort))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", host, hostPort.Port()), nil
}

func (c *TContainer) UnexposedEndpoint(ctx context.Context) (string, error) {
	ipC, err := c.ContainerIP(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", ipC, c.ContainerPort), nil
}

func WithDockerFile(cmd string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.FromDockerfile = testcontainers.FromDockerfile{
			Context:       "../../../",
			Dockerfile:    fmt.Sprintf("./docker/%s.dockerfile", cmd),
			PrintBuildLog: true,
		}
	}
}

func WithWaitStrategy(strategies ...wait.Strategy) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(1 * time.Minute)
	}
}

func WithPort(port string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		if req.ExposedPorts == nil {
			req.ExposedPorts = []string{}
		}
		req.ExposedPorts = append(req.ExposedPorts, port)
	}
}

func WithEnv(envs map[string]string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		if req.Env == nil {
			req.Env = make(map[string]string)
		}
		for k, v := range envs {
			req.Env[k] = v
		}
	}
}

func WithFiles(files map[string]string) func(req *testcontainers.ContainerRequest) {
	cfs := make([]testcontainers.ContainerFile, 0, len(files))
	for hostPath, containerPath := range files {
		cfs = append(cfs, testcontainers.ContainerFile{
			HostFilePath:      hostPath,
			ContainerFilePath: containerPath,
			FileMode:          0744,
		})
	}
	return func(req *testcontainers.ContainerRequest) {
		req.Files = cfs
	}
}

func WithNetwork(network string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Networks = []string{network}
	}
}
