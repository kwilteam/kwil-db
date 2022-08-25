package util

//goland:noinspection SpellCheckingInspection
import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/ignite/cli/ignite/pkg/cosmoscmd"
	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/app"
	"github.com/kwilteam/kwil-db/internal/utils"
	"github.com/spf13/cobra"
)

// The below is for dev mode and makes certain assumptions about
// config options (read through code for more details)

var loggingEnabled = false
var hasReset = false
var hasRun = false
var homeDir string
var homeDirRoot string

func init() {
	os.Args, loggingEnabled = hasFlagAndRemove(os.Args, "--dbg-log-enabled")
	os.Args, hasReset = hasFlagAndRemove(os.Args, "--dbg-reset")
	os.Args, hasRun = hasFlagAndRemove(os.Args, "--dbg-run")

	if !hasReset && !hasRun {
		return
	}

	home := getArg(os.Args, "--home")
	home, exp := utils.TryExpandHomeDir(home)
	if exp {
		replaceArg(os.Args, "--home", home)
	}

	if !strings.HasSuffix(home, "/.kwildb/chain") {
		panic("--home must end with /.kwildb/chain (invalid home directory for dbg mode)")
	}

	homeDirRoot = strings.TrimSuffix(home, "/chain")

	err := os.MkdirAll(home, 0755)
	utils.PanicIfError(err)
}

// BuildAndRunRootCommand When in DEBUG mode, Need to run export periodically since the
// debugger does not allopw capture SIGTERM or SIGINT
// --> os.Args = []string{"kwil-dbd","export","--home", homeDir}
func BuildAndRunRootCommand() error {
	if hasReset {
		// Use "ignite chain serve -r" to reset if this is failing
		return doReset()
	}

	if hasRun {
		return doRun()
	}

	_ = logMessage(os.Args, func(args []string) string {
		return fmt.Sprintf("os.Args = []string{%s}\n", "\""+strings.Join(args, "\",\"")+"\"")
	})

	return svrcmd.Execute(getDefaultRootCommand(), app.DefaultNodeHome)
}

func getHomeDir() (string, error) {
	if homeDir != "" {
		return homeDir, nil
	}

	home := getArg(os.Args, "--home")
	if !strings.HasSuffix(home, "/.kwildb/chain") {
		return "", errors.New("--home must end with /.kwildb/chain (invalid home directory for dbg mode)")
	}

	_ = os.MkdirAll(home, 0755)

	homeDir = home

	return homeDir, nil
}

func getHomeRootDir() (string, error) {
	if homeDirRoot != "" {
		return homeDirRoot, nil
	}

	home, err := getHomeDir()
	if err != nil {
		return "", err
	}

	homeDirRoot = strings.TrimSuffix(home, "/chain")

	return homeDirRoot, nil
}

func getDefaultRootCommand() *cobra.Command {
	rootCmd, _ := cosmoscmd.NewRootCmd(
		app.Name,
		app.AccountAddressPrefix,
		app.DefaultNodeHome,
		app.Name,
		app.ModuleBasics,
		app.New,
		// this line is used by starport scaffolding # root/arguments
	)

	rootCmd.Use = "kwild"
	rootCmd.Short = "kwil-db cli"

	return rootCmd
}

func removeArg(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func getArg(s []string, r string) string {
	for i, v := range s {
		if v == r && i < len(s)-1 {
			return s[i+1]
		}
	}
	return ""
}

func replaceArg(s []string, f, r string) {
	for i, e := range s {
		if e == f {
			if i < len(s)-1 {
				s[i+1] = r
				break
			}
		}
	}
}

func hasFlag(s []string, r string) bool {
	return utils.Any(r, s...)
}

func hasFlagAndRemove(s []string, r string) ([]string, bool) {
	if !hasFlag(s, r) {
		return s, false
	}

	return removeArg(s, r), true
}

func runBlockingDefaultCommand(preBlockC func()) error {
	cmd := getDefaultRootCommand()
	if err := svrcmd.Execute(cmd, app.DefaultNodeHome); err != nil {
		return err
	}

	preBlockC()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	return nil
}

func doRun() error {
	homeDir, err := getHomeDir()
	if err != nil {
		return err
	}

	fn := func() {
		//no-op
	}

	_, err = execute([]string{"export", "--home", homeDir})
	if err != nil {
		fn = func() {
			execute([]string{"export", "--home", homeDir})
		}
	}

	os.Args = []string{"kwild", "start", "--pruning", "nothing", "--grpc.address", "0.0.0.0:9090", "--home", homeDir}
	if err := runBlockingDefaultCommand(fn); err != nil {
		return err
	}

	return nil
}

//goland:noinspection GoUnhandledErrorResult
func doReset() error {
	homeDir, err := getHomeDir()
	if err != nil {
		return err
	}

	os.RemoveAll(homeDir)

	_, err = execute([]string{"init", "mynode", "--chain-id", "kwildb", "--home", homeDir})
	if err != nil {
		return err
	}

	out, err := execute([]string{"keys", "add", "alice", "--output", "text", "--keyring-backend", "test", "--home", homeDir})
	if err != nil {
		return err
	}

	err = addGenesisAccount(homeDir, out)
	if err != nil {
		return err
	}

	out, err = execute([]string{"keys", "add", "bob", "--output", "text", "--keyring-backend", "test", "--home", homeDir})
	if err != nil {
		return err
	}

	err = addGenesisAccount(homeDir, out)
	if err != nil {
		return err
	}

	_, err = execute([]string{"gentx", "alice", "100000000stake", "--chain-id", "kwildb", "--keyring-backend", "test", "--home", homeDir})
	if err != nil {
		return err
	}

	// This seems to sometimes error out
	_, err = execute([]string{"collect-gentxs", "--home", homeDir})
	if err != nil {
		return err
	}

	// An error here is compensated on the next start
	execute([]string{"export", "--home", homeDir})

	os.Args = []string{"kwild", "start", "--pruning", "nothing", "--grpc.address", "0.0.0.0:9090", "--home", homeDir}
	if err := runBlockingDefaultCommand(func() {}); err != nil {
		return err
	}

	return nil
}

func addGenesisAccount(homeDir string, keyring string) error {
	idx := strings.Index(keyring, "kaddr-")
	if idx == -1 {
		return nil
	}

	key := strings.Split(keyring[idx:], "\n")[0]
	key = strings.TrimSpace(key)

	_ = logMessage(key, func(o string) string {
		return o + "\n"
	})

	_, err := execute([]string{"add-genesis-account", key, "20000token,200000000stake", "--home", homeDir})
	if err != nil {
		return err
	}

	return nil
}

func execute(args []string) (string, error) {
	appName, err := os.Executable()
	if err != nil {
		return "", err
	}

	if strings.HasSuffix(appName, "/__debug_bin") {
		appName = appName[:len(appName)-len("/__debug_bin")] + "/kwild"
	} else if strings.HasPrefix(appName, "/tmp/GoLand") {
		dir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		appName = dir + "/cmd/kwil-cosmos/cmd/kwild/kwild"
		if !utils.FileExists(appName) {
			return "", fmt.Errorf("file %s does not exists. Build kwild in order for the shell usage of command delegation", appName)
		}
	} else {
		appName, err = filepath.EvalSymlinks(appName)
		if err != nil {
			return "", err
		}
	}

	cmd := exec.Command(appName, args...)

	b, err := cmd.Output()
	if err != nil {
		_ = logMessage(err, func(e error) string {
			return fmt.Sprintf("execute() error: %s\n", e)
		})

		return "", err
	}

	return string(b), nil
}

func logMessage[T any](arg T, fn func(a T) string) error {
	if !loggingEnabled {
		return nil
	}

	dir, err := getHomeRootDir()
	if err != nil {
		return err
	}

	dir = path.Join(dir, "tmp")
	_ = os.MkdirAll(dir, 0755)

	file := path.Join(dir, "kwild.dbg")

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Unable to create/append 'tmp/kwild.log': %v", err)
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.WriteString(fn(arg))
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		log.Fatalf("Unable to Sync to 'tmp/kwild.log': %v", err)
	}

	return nil
}
