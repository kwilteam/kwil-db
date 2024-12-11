package rpc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/app/shared/bind"
	adminclient "github.com/kwilteam/kwil-db/node/admin"
	"github.com/spf13/cobra"
)

const (
	adminCertName = "admin.cert" // fallback in case we're on a remote machine not kwild's
)

// BindRPCFlags binds the RPC flags to the given command.
// This includes an rpcserver flag, and the TLS flags.
// These flags can be used to create an admin service client.
// The flags will be bound to all subcommands of the given command.
func BindRPCFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("rpcserver", "s", "/tmp/kwild.socket", "admin RPC server address (either unix or tcp)")

	cmd.PersistentFlags().String("authrpc-cert", "", "kwild's TLS certificate, required for HTTPS server")
	cmd.PersistentFlags().String("pass", "", "admin server password (alternative to mTLS with tlskey/tlscert). May be set in ~/.kwil-admin/rpc-admin-pass instead.")
	cmd.PersistentFlags().String("tlskey", "auth.key", "kwil-admin's TLS key file to establish a mTLS (authenticated) connection")
	cmd.PersistentFlags().String("tlscert", "auth.cert", "kwil-admin's TLS certificate file for server to authenticate us")
}

// GetRPCServerFlag returns the RPC flag from the given command.
func GetRPCServerFlag(cmd *cobra.Command) (string, error) {
	return cmd.Flags().GetString("rpcserver")
}

// AdminSvcClient will return an admin service client based on the flags.
// The flags should be bound using the BindRPCFlags function.
func AdminSvcClient(ctx context.Context, cmd *cobra.Command) (*adminclient.AdminClient, error) {
	adminOpts := []adminclient.Opt{}

	rpcServer, err := GetRPCServerFlag(cmd)
	if err != nil {
		return nil, err
	}

	// get the tls files
	// if one is specified, all must be specified
	// if none are specified, then we do not use tls
	if cmd.Flags().Changed("authrpc-cert") || cmd.Flags().Changed("tlskey") || cmd.Flags().Changed("tlscert") {
		kwildTLSCertFile, clientTLSKeyFile, clientTLSCertFile, err := getTLSFlags(cmd)
		if err != nil {
			return nil, err
		}

		adminOpts = append(adminOpts, adminclient.WithTLS(kwildTLSCertFile, clientTLSKeyFile, clientTLSCertFile))
	}

	if pass, err := cmd.Flags().GetString("pass"); err != nil {
		return nil, err
	} else if pass != "" {
		adminOpts = append(adminOpts, adminclient.WithPass(pass))
	}

	return adminclient.NewClient(ctx, rpcServer, adminOpts...)
}

// getTLSFlags returns the TLS flags from the given command.
func getTLSFlags(cmd *cobra.Command) (kwildTLSCertFile, clientTLSKeyFile, clientTLSCertFile string, err error) {
	kwildTLSCertFile, err = cmd.Flags().GetString("authrpc-cert")
	if err != nil {
		return "", "", "", err
	}

	clientTLSKeyFile, err = cmd.Flags().GetString("tlskey")
	if err != nil {
		return "", "", "", err
	}

	clientTLSCertFile, err = cmd.Flags().GetString("tlscert")
	if err != nil {
		return "", "", "", err
	}

	cert := nodeCert{
		KwildTLSCertFile:  kwildTLSCertFile,
		ClientTLSKeyFile:  clientTLSKeyFile,
		ClientTLSCertFile: clientTLSCertFile,
	}

	rootDir, err := bind.RootDir(cmd)
	if err != nil {
		return "", "", "", err
	}

	return cert.tlsFiles(rootDir)
}

// nodeCert is the struct that holds the TLS certificate and key files for
// kwil-admin to use when connecting to the remote node.
type nodeCert struct {
	KwildTLSCertFile  string
	ClientTLSKeyFile  string // default: auth.key
	ClientTLSCertFile string // default: auth.cert
}

// tlsFiles loads the remote nodes TLS certificate, which the client uses to
// authenticate the server during the connection handshake, and our TLS key
// pair, which allows the server to authenticate the client. The server's TLS
// certificate would be obtained from the node machine and specified with the
// --authrpc-cert flag, this will search for it in known paths. The client key
// pair (ours) should first be generated with the `node gen-auth-key` command.
func (nc *nodeCert) tlsFiles(rootDir string) (nodeCert, ourKey, ourCert string, err error) {
	// Look for kwild's TLS certificate in:
	//  1. any path provided via --authrpc-cert
	//  2. ~/.kwil-admin/kwild.cert
	//  3. ~/.kwild/rpc.cert
	nodeCert = nc.KwildTLSCertFile
	if nodeCert != "" { // --authrpc-cert
		nodeCert = fullPath(nodeCert, rootDir)
	} else { // search the two fallback paths
		nodeCert = fullPath(adminCertName, rootDir)
	}
	if nodeCert == "" || !fileExists(nodeCert) {
		err = fmt.Errorf("kwild cert file not found, checked %v", nodeCert)
		return
	}
	ourKey, ourCert = fullPath(nc.ClientTLSKeyFile, rootDir), fullPath(nc.ClientTLSCertFile, rootDir)
	if ourKey == "" || ourCert == "" {
		err = errors.New("our TLS key/cert not found")
		return // leave the existence check until we load the files
	}

	return
}

// fullPath gets the full path to a file, searching in the following order:
//  1. the path itself, if it is absolute
//  2. the current directory
//  3. provided root directory
func fullPath(path, rootDir string) string {
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

	// Check for the file name in root dir
	fullPath = filepath.Join(rootDir, path)
	if fileExists(fullPath) {
		return fullPath
	}

	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
