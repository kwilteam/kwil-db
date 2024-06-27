package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

// EnsureNetworkExist creates a new docker network with a random UUID name.
func EnsureNetworkExist(ctx context.Context, testName string) (
	*testcontainers.DockerNetwork, error) {
	net, err := network.New(ctx,
		network.WithCheckDuplicate(),
		network.WithAttachable(),
		//network.WithInternal(), // we need to expose the network to the host
		network.WithLabels(map[string]string{"test": "integration"}),
		network.WithDriver("bridge"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create docker network for %s: %w", testName, err)
	} else {
		fmt.Printf("docker network created: %s(%s) for %s\n", net.Name, net.ID, testName)
		return net, nil
	}
}

// getEndpoints returns proto://host:port exposed/unexposed string for given exposed port.
func getEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context,
	port nat.Port, proto string) (exposed string, unexposed string, err error) {
	exposed, err = ctr.PortEndpoint(ctx, port, proto)
	if err != nil {
		return
	}

	ctrInspect, err := ctr.Inspect(ctx)
	if err != nil {
		return
	}

	// NOTE: sometime the container name is returned with leading slash
	ctrName := strings.TrimPrefix(ctrInspect.Name, "/")
	unexposed = fmt.Sprintf("%s://%s:%s", proto, ctrName, port.Port())
	return
}

func ETHDevNetHTTPEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "8545", "http")
}

func ETHDevNetWSEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "8545", "ws")
}

func KwildJSONRPCEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "8484", "http")
}

func KwildHTTPEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "8080", "http")
}

func KwildWSEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "8080", "ws")
}

func KwildAdminEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "8485", "tcp") // unused because we are using kwil-admin inside the container with a unix socket
}

func KwildP2pEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "26656", "tcp")
}

func KwildRpcEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "26657", "http")
}

func KwildDebugEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "40000", "tcp")
}

func PostgresEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "5432", "tcp")
}

func MathExtEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "50051", "http")
}
