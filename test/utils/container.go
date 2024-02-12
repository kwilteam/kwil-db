package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

func EnsureNetworkExist(ctx context.Context, networkName string) (
	*testcontainers.DockerNetwork, error) {
	net, err := newNetwork(ctx,
		networkName,
		network.WithCheckDuplicate(),
		network.WithAttachable(),
		network.WithInternal(),
		network.WithLabels(map[string]string{"test": "integration"}),
	)

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Printf("docker network %s already exists\n", networkName)
			return nil, nil
		} else {
			return nil, fmt.Errorf("failed to create docker network %s: %w", networkName, err)
		}
	} else {
		fmt.Printf("docker network created: %s(%s)\n", net.Name, net.ID)
		return net, nil
	}
}

// newNetwork is basically the same as network.New, but it supports create with
// a specific name. No idea why the original network.New doesn't support this.
func newNetwork(ctx context.Context, name string,
	opts ...network.NetworkCustomizer) (*testcontainers.DockerNetwork, error) {
	nc := types.NetworkCreate{
		Driver: "bridge",
		Labels: testcontainers.GenericLabels(),
	}

	for _, opt := range opts {
		opt.Customize(&nc)
	}

	//nolint:staticcheck
	netReq := testcontainers.NetworkRequest{
		Driver:         nc.Driver,
		CheckDuplicate: nc.CheckDuplicate,
		Internal:       false,
		EnableIPv6:     nc.EnableIPv6,
		Name:           name,
		Labels:         nc.Labels,
		Attachable:     nc.Attachable,
		IPAM:           nc.IPAM,
	}

	//nolint:staticcheck
	n, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: netReq,
	})
	if err != nil {
		return nil, err
	}

	// Return a DockerNetwork struct instead of the Network interface,
	// following the "accept interface, return struct" pattern.
	return n.(*testcontainers.DockerNetwork), nil
}

// getEndpoints returns proto://host:port exposed/unexposed string for given exposed port.
func getEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context,
	port nat.Port, proto string) (exposed string, unexposed string, err error) {
	exposed, err = ctr.PortEndpoint(ctx, port, proto)
	if err != nil {
		return
	}

	ctrName, err := ctr.Name(ctx)
	if err != nil {
		return
	}

	// NOTE: sometime the container name is returned with leading slash
	if strings.HasPrefix(ctrName, "/") {
		ctrName = strings.TrimPrefix(ctrName, "/")
	}
	unexposed = fmt.Sprintf("%s://%s:%s", proto, ctrName, port.Port())
	return
}

func GanacheHTTPEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "8545", "http")
}

func GanacheWSEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "8545", "ws")
}

func KwildGRPCEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "50051", "tcp")
}

func KwildHTTPEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "8080", "http")
}

func KwildWSEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "8080", "ws")
}

func KwildAdminEndpoints(ctr *testcontainers.DockerContainer, ctx context.Context) (string, string, error) {
	return getEndpoints(ctr, ctx, "50151", "tcp")
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
