package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"runtime"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/internal/pkg/build"
	"github.com/spf13/cobra"
	"github.com/tonistiigi/go-rosetta"
)

var versionTemplate = `
 Version:	{{.Version}}
 Git commit:	{{.GitCommit}}
 Built:	{{.BuildTime}}
 API version:	{{.APIVersion}}
 Go version:	{{.GoVersion}}
 OS/Arch:	{{.Os}}/{{.Arch}}`

type versionInfo struct {
	// build-time info
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	// client machine info
	APIVersion string `json:"api_version"`
	GoVersion  string `json:"go_version"`
	Os         string `json:"os"`
	Arch       string `json:"arch"`
}

type respVersionInfo struct {
	Info *versionInfo
}

func (v *respVersionInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Info)
}

func (v *respVersionInfo) MarshalText() ([]byte, error) {
	tmpl := template.New("version")
	// load different template according to the opts.format
	tmpl, err := tmpl.Parse(versionTemplate)
	if err != nil {
		return []byte(""), fmt.Errorf("template parsing error: %w", err)
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, v.Info)
	if err != nil {
		return []byte(""), fmt.Errorf("template executing error: %w", err)
	}

	bs, err := io.ReadAll(&buf)
	if err != nil {
		return []byte(""), err
	}

	return bs, nil
}

func NewVersionCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version [OPTIONS]",
		Short: "Show the kwil-cli version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resp := &respVersionInfo{
				Info: &versionInfo{
					Version:    build.Version,
					APIVersion: "",
					GitCommit:  build.GitCommit,
					GoVersion:  runtime.Version(),
					Os:         runtime.GOOS,
					Arch:       arch(),
					BuildTime:  build.BuildTime,
				},
			}

			return display.Print(resp, nil, config.GetOutputFormat())
		},
	}

	return cmd
}

func arch() string {
	arch := runtime.GOARCH
	if rosetta.Enabled() {
		arch += " (rosetta)"
	}
	return arch
}
