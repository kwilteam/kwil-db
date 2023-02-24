package system

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tonistiigi/go-rosetta"
	"html/template"
	"kwil/internal/pkg/build"
	"os"
	"runtime"
	"text/tabwriter"
)

var versionTemplate = `
 Version:	{{.Version}}
 Git commit:	{{.GitCommit}}
 Built:	{{.BuildTime}}
 API version:	{{.APIVersion}}
 Go version:	{{.GoVersion}}
 OS/Arch:	{{.Os}}/{{.Arch}}`

type versionOptions struct {
	format string
}

type versionInfo struct {
	// build-time info
	Version   string
	GitCommit string
	BuildTime string `json:",omitempty"`
	// client machine info
	APIVersion string `json:"ApiVersion"`
	GoVersion  string
	Os         string
	Arch       string
}

func NewVersionCmd() *cobra.Command {
	var opts versionOptions

	var cmd = &cobra.Command{
		Use:   "version [OPTIONS]",
		Short: "Show the kwil-cli version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVesrion(&opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.format, "format", "f", "", "Format the output using the given Go template")

	return cmd
}

func arch() string {
	arch := runtime.GOARCH
	if rosetta.Enabled() {
		arch += " (rosetta)"
	}
	return arch
}

func runVesrion(opts *versionOptions) error {
	tmpl := template.New("version")
	tmpl, err := tmpl.Parse(versionTemplate)
	if err != nil {
		return errors.Wrap(err, "template parsing error")
	}

	vd := versionInfo{
		Version:    build.Version,
		APIVersion: "",
		GitCommit:  build.GitCommit,
		GoVersion:  runtime.Version(),
		Os:         runtime.GOOS,
		Arch:       arch(),
		BuildTime:  build.BuildTime,
	}

	// @yaiba TODO: add server version?
	return prettyPrintVersion(vd, tmpl)
}

func prettyPrintVersion(vd versionInfo, tmpl *template.Template) error {
	t := tabwriter.NewWriter(os.Stdout, 20, 1, 1, ' ', 0)
	err := tmpl.Execute(t, vd)
	_, _ = t.Write([]byte("\n"))
	t.Flush()
	return err
}
