package adapters

import (
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

//func init() {
//	os.Setenv("DOCKER_HOST", fmt.Sprintf("unix://%s/.docker/run/docker.sock", os.Getenv("HOME")))
//}

const (
	kwildTestNetworkName = "kwild-test-network"
	startupTimeout       = 1 * time.Minute
)

type containerOption func(req *testcontainers.ContainerRequest)

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
