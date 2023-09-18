package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/internal/pkg/transport"
	"github.com/kwilteam/kwil-db/pkg/admin"
)

const (
	// kwildTLSCertFileName is the default file name for kwild's TLS certificate
	// when we look in the .kwild folder for it (when kwild is on this machine).
	kwildTLSCertFileName = config.DefaultTLSCertFile

	// capturedKwildTLSCertFileName is the name of the kwild TLS certificate when
	// we look in the .kwil-admin folder for it.
	capturedKwildTLSCertFileName = "kwild.cert"
)

// kwil-admin node ... (running node administration: ping, peer mgmt, sql query, )

type NodeCmd struct {
	RPCServer string `arg:"-s,--rpcserver" default:"127.0.0.1:50151" help:"admin RPC server address"`

	KwildTLSCertFile  string `arg:"--authrpc-cert" help:"kwild's TLS certificate"`
	ClientTLSKeyFile  string `arg:"--tlskey" default:"auth.key" help:"kwil-admin's TLS key file to establish a mTLS (authenticated) connection"`
	ClientTLSCertFile string `arg:"--tlscert" default:"auth.cert" help:"kwil-admin's TLS certificate file for server to authenticate us"`

	Ping    *NodePingCmd       `arg:"subcommand:ping" help:"Check connectivity with the node's admin RPC interface"`
	Version *NodeVerCmd        `arg:"subcommand:version" help:"Report the version of the remote node"`
	Status  *NodeStatusCmd     `arg:"subcommand:status" help:"Show detailed node status"`
	Peers   *NodePeersCmd      `arg:"subcommand:peers" help:"Show info on all current peers"`
	GenKey  *NodeGenAuthKeyCmd `arg:"subcommand:gen-auth-key" help:"Generate a client key pair."`
}

type NodeGenAuthKeyCmd struct{}

func (ngkc *NodeGenAuthKeyCmd) run(ctx context.Context, a *args) error {
	rootDir := defaultKwilAdminRoot()

	clientKeyFile := a.Node.ClientTLSKeyFile
	if !filepath.IsAbs(clientKeyFile) {
		clientKeyFile = filepath.Join(rootDir, clientKeyFile)
	}
	if fileExists(clientKeyFile) {
		return fmt.Errorf("key file exists: %v", clientKeyFile)
	}

	clientCertFile := a.Node.ClientTLSCertFile
	if !filepath.IsAbs(clientCertFile) {
		clientCertFile = filepath.Join(rootDir, clientCertFile)
	}
	if fileExists(clientCertFile) {
		return fmt.Errorf("cert file exists: %v", clientCertFile)
	}

	return transport.GenTLSKeyPair(clientCertFile, clientKeyFile, "kwild CA", nil)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (*NodeCmd) fullPath(path string) string {
	// If an absolute path is specified, do nothing.
	if filepath.IsAbs(path) {
		return path
	}

	// First check relative to the current directory.
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	if fileExists(fullPath) {
		return fullPath
	}

	// Check for the file name in ~/.kwil-admin
	fullPath = filepath.Join(defaultKwilAdminRoot(), path)
	if fileExists(fullPath) {
		return fullPath
	}

	return ""
}

// tlsFiles loads the remote nodes TLS certificate, which the client uses to
// authenticate the server during the connection handshake, and our TLS key
// pair, which allows the server to authenticate the client. The server's TLS
// certificate would be obtained from the node machine and specified with the
// --authrpc-cert flag, this will search for it in known paths. The client key
// pair (ours) should first be generated with the `node gen-auth-key` command.
func (nc *NodeCmd) tlsFiles() (nodeCert, ourKey, ourCert string, err error) {
	// Look for kwild's TLS certificate in:
	//  1. any path provided via --authrpc-cert
	//  2. ~/.kwil-admin/kwild.cert
	//  3. ~/.kwild/rpc.cert
	nodeCert = nc.KwildTLSCertFile
	if nodeCert != "" { // --authrpc-cert
		nodeCert = nc.fullPath(nodeCert)
	} else { // search the two fallback paths
		nodeCert = filepath.Join(defaultKwilAdminRoot(), capturedKwildTLSCertFileName)
		if !fileExists(nodeCert) {
			nodeCert = filepath.Join(defaultKwildRoot(), kwildTLSCertFileName)
		}
	}
	if nodeCert == "" || !fileExists(nodeCert) {
		err = fmt.Errorf("kwild cert file not found, checked %v", nodeCert)
		return
	}
	ourKey, ourCert = nc.fullPath(nc.ClientTLSKeyFile), nc.fullPath(nc.ClientTLSCertFile)
	if ourKey == "" || ourCert == "" {
		err = errors.New("our TLS key/cert not found")
		return // leave the existence check until we load the files
	}

	return
}

func (nc *NodeCmd) newClient() (*admin.Client, error) {
	kwildCert, ourKey, ourCert, err := nc.tlsFiles()
	if err != nil {
		return nil, err
	}
	return admin.New(nc.RPCServer, kwildCert, ourKey, ourCert)
}

type NodePingCmd struct{}

func (pc *NodePingCmd) run(ctx context.Context, a *args) error {
	client, err := a.Node.newClient()
	if err != nil {
		return err
	}
	ping, err := client.Ping(ctx)
	if err != nil {
		return err
	}
	fmt.Println(ping)
	return nil
}

type NodeVerCmd struct{}

func (pc *NodeVerCmd) run(ctx context.Context, a *args) error {
	client, err := a.Node.newClient()
	if err != nil {
		return err
	}
	ver, err := client.Version(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("kwild version %v\n", ver)
	return nil
}

type NodeStatusCmd struct{}

func (nsc *NodeStatusCmd) run(ctx context.Context, a *args) error {
	client, err := a.Node.newClient()
	if err != nil {
		return err
	}
	status, err := client.Status(ctx)
	if err != nil {
		return err
	}
	statusJSON, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(statusJSON))
	return nil
}

type NodePeersCmd struct{}

func (nsc *NodePeersCmd) run(ctx context.Context, a *args) error {
	client, err := a.Node.newClient()
	if err != nil {
		return err
	}
	peers, err := client.Peers(ctx)
	if err != nil {
		return err
	}
	peersJSON, err := json.MarshalIndent(peers, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(peersJSON))
	return nil
}
