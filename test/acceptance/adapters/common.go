package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//func init() {
//	os.Setenv("DOCKER_HOST", fmt.Sprintf("unix://%s/.docker/run/docker.sock", os.Getenv("HOME")))
//}

const (
	kwilTestNetworkName = "kwil-test-network"
	startupTimeout      = 15 * time.Second
	waitTimeout         = 15 * time.Second
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
		return fmt.Errorf("could not get exposed endpoint: %w", err)
	}

	unexposedEndpoint, err := c.UnexposedEndpoint(ctx)
	if err != nil {
		return fmt.Errorf("could not get unexposed endpoint: %w", err)
	}

	name, err := c.Name(ctx)
	if err != nil {
		return fmt.Errorf("could not get container name: %w", err)
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
			Context:       "../../",
			Dockerfile:    fmt.Sprintf("./docker/%s.dockerfile", cmd),
			PrintBuildLog: true,
		}
	}
}

func WithWaitStrategy(strategies ...wait.Strategy) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(waitTimeout)
	}
}

func WithExposedPort(port string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		if req.ExposedPorts == nil {
			req.ExposedPorts = []string{}
		}
		req.ExposedPorts = append(req.ExposedPorts, port)
	}
}

func WithExposedPorts(ports []string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		if req.ExposedPorts == nil {
			req.ExposedPorts = []string{}
		}
		req.ExposedPorts = append(req.ExposedPorts, ports...)
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
