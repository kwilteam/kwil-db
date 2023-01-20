package grpcserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"kwil/cmd/kwil-cli/common"
	grpc_client "kwil/kwil/client/grpc-client"
	"kwil/tests/integration/adapters"
	"kwil/tests/integration/specifications"
	"kwil/x/types/databases"
	"kwil/x/types/transactions"
	"os"
	"path/filepath"
	"testing"

	anytype "kwil/x/types/data_types/any_type"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

func loadDatabase(path string) (*databases.Database[anytype.KwilAny], error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var db databases.Database[anytype.KwilAny]
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, err
	}

	return &db, nil
}

func TestKwilDatabase(t *testing.T) {
	ctx := context.Background()

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type testCase struct {
		name    string
		args    *databases.Database[anytype.KwilAny]
		wantErr bool
		want    *transactions.Response
	}

	tests := []testCase{}
	loadCases := func(_dir string, positive bool) {
		filepath.Walk(_dir, func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			db, err := loadDatabase(path)
			if err != nil {
				t.Fatalf("load test case failed: %s", err)
			}
			tests = append(tests, testCase{
				name:    path,
				args:    db,
				wantErr: !positive,
			})
			return nil
		})
	}
	loadCases("./data/positive-cases", true)
	loadCases("./data/negative-cases", false)

	// start the docker container
	common.LoadConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := adapters.StartDBDockerService(t, ctx)
			kc := adapters.StartKwildDockerService(t, ctx, dc)
			ipC, err := kc.ContainerIP(ctx)
			assert.NoError(t, err)
			portC, err := kc.MappedPort(ctx, nat.Port(kc.Port))
			assert.NoError(t, err)
			t.Logf("create driver to %s", fmt.Sprintf("%s:%s", ipC, portC.Port()))
			driver := grpc_client.Driver{Addr: fmt.Sprintf("%s:%s", ipC, portC.Port())}
			//kc.Terminate(ctx)
			//dc.Terminate(ctx)

			specifications.DeploySpecification(t, &driver, tt.args, tt.wantErr, tt.want)
		})
	}

}
