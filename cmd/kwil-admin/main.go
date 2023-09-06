package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/kwilteam/kwil-db/internal/pkg/version"

	"github.com/alexflint/go-arg"
)

const (
	exitSuccess = iota
	exitAppError
	exitBadArgs
)

type args struct {
	Key   *KeyCmd        `arg:"subcommand:key"`
	Setup *SetupCmd      `arg:"subcommand:setup"`
	Vals  *ValidatorsCmd `arg:"subcommand:validators"`
	Node  *NodeCmd       `arg:"subcommand:node"`

	Help *HelpCmd `arg:"subcommand:help"`
	Ver  *VerCmd  `arg:"subcommand:version"`
}

func defaultRoot() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kwild")
}

// HelpCmd and VerCmd are defined to emulate -h,--help and --version.
type HelpCmd struct{}
type VerCmd struct{}

func (*args) Version() string {
	return "kwil-admin version " + version.KwilVersion
}

func (*args) Description() string {
	return "kwil-admin: The Kwil node admin tool\n"
}

func (*args) Epilogue() string {
	return "For more information visit https://docs.kwil.com/"
}

// runner may be implemented by any subcommand to be a directly-runnable
// subcommand that receives the full args struct rather than using the full
// cascade of switches through the subcommands. EXPERIMENT.
type runner interface {
	run(context.Context, *args) error
}

// run is a fallback dispatcher if the subcommand is not a runner.
func (a *args) run(ctx context.Context) error {
	switch {
	case a.Key != nil:
		return a.Key.run(ctx)
	case a.Setup != nil:
		return a.Setup.run(ctx)
	case a.Node != nil:
		return a.Node.run(ctx)
	// case a.Vals != nil: // made these runners to test the pattern
	// 	return a.Vals.run(ctx)
	case a.Ver != nil:
		fmt.Fprintln(os.Stdout, a.Version()) // emulate --version
		return nil
	default:
		return arg.ErrHelp
	}
}

func main() {
	a := &args{}
	p := arg.MustParse(a) // parse and maybe show help or version and exit
	if p.Subcommand() == nil {
		p.WriteHelp(os.Stderr)
		fmt.Println("no command given")
		os.Exit(exitBadArgs)
	}
	if _, isHelp := p.Subcommand().(*HelpCmd); isHelp {
		p.WriteHelpForSubcommand(os.Stdout) // emulate -h,--help
		os.Exit(exitSuccess)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-signalChan
		cancel()
	}()

	// EXPERIMENT: directly-runnable subcommands to reduce boilerplate switches
	var err error
	r, ok := p.Subcommand().(runner)
	if ok {
		err = r.run(ctx, a)
	} else {
		// not a runnable subcommand, use switch to dispatch
		err = a.run(ctx)
	}
	if errors.Is(err, arg.ErrHelp) {
		p.WriteHelpForSubcommand(os.Stderr, p.SubcommandNames()...)
		os.Exit(exitBadArgs)
	}
	if err != nil {
		p.FailSubcommand(err.Error(), p.SubcommandNames()...)
		os.Exit(exitAppError)
	}

	os.Exit(exitSuccess)
}
