package node

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/core/rpc/transport"
	"github.com/spf13/cobra"
)

var (
	genAuthKeyLong = `Generate a new key pair for use in the node's validator set.
	
The key pair is generated and stored in the node's configuration directory, in the files auth.key and auth.cert. The key pair is used to authenticate the admin tool to the node.`
	genAuthKeyExample = `# Generate a new TLS key pair to talk to the node
kwil-admin node gen-auth-key --authrpc-cert "~/.kwild/rpc.cert"`
)

func genAuthKeyCmd() *cobra.Command {
	var keyFile, certFile string

	cmd := &cobra.Command{
		Use:     "gen-auth-key",
		Short:   "Generate a new key pair for use in the node's validator set.",
		Long:    genAuthKeyLong,
		Example: genAuthKeyExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir := common.DefaultKwilAdminRoot()

			if !filepath.IsAbs(keyFile) {
				keyFile = filepath.Join(rootDir, keyFile)
			}
			if fileExists(keyFile) {
				return display.PrintErr(cmd, fmt.Errorf("key file exists: %v", keyFile))
			}
			if err := os.MkdirAll(filepath.Dir(keyFile), 0755); err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to create key file dir: %v", err))
			}

			if !filepath.IsAbs(certFile) {
				certFile = filepath.Join(rootDir, certFile)
			}
			if fileExists(certFile) {
				return display.PrintErr(cmd, fmt.Errorf("cert file exists: %v", certFile))
			}
			if err := os.MkdirAll(filepath.Dir(certFile), 0755); err != nil {

				return display.PrintErr(cmd, fmt.Errorf("failed to create key file dir: %v", err))
			}

			return transport.GenTLSKeyPair(certFile, keyFile, "kwild CA", nil)
		},
	}

	cmd.Flags().StringVar(&keyFile, "tlskey", "auth.key", "output path for the node's TLS key file")
	cmd.Flags().StringVar(&certFile, "tlscert", "auth.cert", "output path for the node's TLS certificate file")

	return cmd
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
