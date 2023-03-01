package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kwil/internal/pkg/graphql/query"
	"kwil/pkg/databases"
	grpc "kwil/pkg/grpc/client"
	"kwil/pkg/log"
	big2 "kwil/pkg/utils/numbers/big"
	"math/big"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

func MustRun(cmd *exec.Cmd) error {
	cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout
	// here we ignore the stdout
	var out bytes.Buffer
	cmd.Stdout = &out
	return cmd.Run()
}

// KwilCliDriver is a cli driver for integration tests
type KwilCliDriver struct {
	cliBin        string
	chainRPCURL   string
	nodeRPCURL    string
	nodeGwURL     string
	pk            string
	walletAddress string
	logger        log.Logger
}

func NewKwilCliDriver(bin, nodeRPCURL, nodeGwUrl, chainRpcUrl, pk, walletAddress string, logger log.Logger) *KwilCliDriver {
	return &KwilCliDriver{
		cliBin:        bin,
		chainRPCURL:   chainRpcUrl,
		nodeRPCURL:    nodeRPCURL,
		nodeGwURL:     nodeGwUrl,
		pk:            pk,
		walletAddress: walletAddress,
		logger:        logger,
	}
}

func (d *KwilCliDriver) newCmd(args ...string) *exec.Cmd {
	//args = append(args, "--client-chain-provider", d.chainRpcUrl)
	//args = append(args, "--kwil-provider", d.nodeUrl)
	//args = append(args, "--private-key", d.pk)
	//fmt.Printf("cmd to exec:  %s, %+v\n", d.cliBin, args)
	cmd := exec.Command(d.cliBin, args...)
	cmd.Env = append(cmd.Environ(), "KCLI_WALLET_PRIVATE_KEY="+d.pk)
	cmd.Env = append(cmd.Environ(), "KCLI_NODE_RPC_URL="+d.nodeRPCURL)
	cmd.Env = append(cmd.Environ(), "KCLI_CHAIN_RPC_URL="+d.chainRPCURL)
	return cmd
}

func (d *KwilCliDriver) GetUserAddress() string {
	return d.walletAddress
}

func (d *KwilCliDriver) GetServiceConfig(ctx context.Context) (svcCfg grpc.SvcConfig, err error) {
	out, err := d.newCmd("utils", "node-config").Output()
	if err != nil {
		return svcCfg, fmt.Errorf("failed to deposit: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if line == "" ||
			strings.Contains(line, "Funding:") ||
			strings.Contains(line, "Gateway:") {
			continue
		}

		if strings.Contains(line, "ChainCode:") {
			code := strings.TrimSpace(strings.Split(line, ":")[1])
			svcCfg.Funding.ChainCode, err = strconv.ParseInt(code, 10, 64)
			if err != nil {
				return
			}
		}
		if strings.Contains(line, "PoolAddress:") {
			svcCfg.Funding.PoolAddress = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.Contains(line, "ProviderAddress:") {
			svcCfg.Funding.ProviderAddress = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.Contains(line, "RpcUrl:") {
			svcCfg.Funding.RpcUrl = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.Contains(line, "GraphqlUrl:") {
			svcCfg.Gateway.GraphqlUrl = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}

	return
}

func (d *KwilCliDriver) DepositFund(ctx context.Context, amount *big.Int) error {
	cmd := d.newCmd("fund", "deposit", amount.String(), "-y")
	err := MustRun(cmd)
	if err != nil {
		return fmt.Errorf("failed to deposit: %w", err)
	}
	return nil
}

func (d *KwilCliDriver) GetDepositBalance(ctx context.Context) (*big.Int, error) {
	cmd := d.newCmd("fund", "balances")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get deposited balance: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Deposit Balance: ") {
			n := strings.TrimPrefix(line, "Deposit Balance: ")
			return big2.BigStr(n).AsBigInt()
		}
	}

	return nil, nil
}

func (d *KwilCliDriver) ApproveToken(ctx context.Context, spender string, amount *big.Int) error {
	// approve cmd will get spender address
	cmd := d.newCmd("fund", "approve", amount.String(), "-y")
	err := MustRun(cmd)
	if err != nil {
		return fmt.Errorf("failed to approve: %w", err)
	}
	return nil
}

func (d *KwilCliDriver) GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error) {
	cmd := d.newCmd("fund", "balances", "--address", from)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get allowance: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Allowance: ") {
			n := strings.TrimPrefix(line, "Allowance: ")
			return big2.BigStr(n).AsBigInt()
		}
	}

	return nil, nil
}

func (d *KwilCliDriver) DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) error {
	schemaFile := path.Join(os.TempDir(), fmt.Sprintf("schema-%s.json", time.Now().Format("20060102150405")))

	dbByte, err := json.Marshal(db)
	if err != nil {
		return fmt.Errorf("failed to marshal database: %w", err)
	}

	err = os.WriteFile(schemaFile, dbByte, 0644)
	if err != nil {
		return fmt.Errorf("failed to write database schema: %w", err)
	}

	cmd := d.newCmd("database", "deploy", "-p", schemaFile)
	err = MustRun(cmd)
	if err != nil {
		return fmt.Errorf("failed to deploy databse: %w", err)
	}
	return nil
}

func (d *KwilCliDriver) DatabaseShouldExists(ctx context.Context, owner string, dbName string) error {
	cmd := d.newCmd("database", "list", "-o", owner)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list database: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, dbName) {
			return nil
		}
	}

	return fmt.Errorf("database does not exist")
}

func (d *KwilCliDriver) ExecuteQuery(ctx context.Context, dbName string, queryName string, queryInputs []string) error {
	args := []string{"database", "execute"}
	args = append(args, queryInputs...)
	args = append(args, "--name", dbName)
	args = append(args, "--query", queryName)
	cmd := d.newCmd(args...)
	err := MustRun(cmd)
	if err != nil {
		return fmt.Errorf("failed to execute database: %w", err)
	}
	return nil
}

func (d *KwilCliDriver) DropDatabase(ctx context.Context, dbName string) error {
	cmd := d.newCmd("database", "drop", dbName)
	err := MustRun(cmd)
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}
	return nil
}

func (d *KwilCliDriver) QueryDatabase(ctx context.Context, queryStr string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/graphql", d.nodeGwURL)
	return query.Query(ctx, url, queryStr)
}
