package utils

import (
	"bytes"
	"os"
	"text/template"
)

const DefaultDockerImage = "kwild:latest"

type ComposeConfig struct {
	Network string
	// ExposedHTTPPorts can be left empty to not expose any ports to the host,
	// or set to the host ports to expose the http interface for each node. e.g.
	// []int{8081, 8082, ...}
	ExposedHTTPPorts []int
	DockerImage      string
}

func genCompose(templateFile string, config ComposeConfig) (string, error) {
	tpt, err := os.ReadFile(templateFile)
	if err != nil {
		return "", err
	}

	if config.DockerImage == "" {
		config.DockerImage = DefaultDockerImage
	}

	funcMap := template.FuncMap{
		"plus": func(i, j int) int {
			return i + j
		},
	}

	tmpl, err := template.New("test-docker-compose").Funcs(funcMap).Parse(string(tpt))
	if err != nil {
		return "", err
	}

	var res bytes.Buffer
	err = tmpl.Execute(&res, config)
	if err != nil {
		return "", err
	}

	return res.String(), nil
}

func CreateComposeFile(targetFile, templateFile string, config ComposeConfig) error {
	content, err := genCompose(templateFile, config)
	if err != nil {
		return err
	}

	return os.WriteFile(targetFile, []byte(content), 0644)
}
