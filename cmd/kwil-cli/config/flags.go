package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	EnvPrefix = "KWIL_CLI"
)

var DefaultConfigFile string

func init() {
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.SetEnvPrefix(EnvPrefix)

	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "/tmp"
	}

	configPath := filepath.Join(dirname, ".kwil_cli")

	DefaultConfigFile = filepath.Join(configPath, "config.json")

	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()
}

// OutputFormat is the format for command output
// It implements the pflag.Value interface
type OutputFormat string

// String implements the Stringer interface
// NOTE: cannot use the pointer receiver here
func (o OutputFormat) String() string {
	return string(o)
}

func (o *OutputFormat) Set(s string) error {
	switch s {
	case "text", "json":
		*o = OutputFormat(s)
		return nil
	default:
		return errors.New(`invalid output format, must be either "text" or "json`)
	}
}

func (o *OutputFormat) Type() string {
	return "output format"
}

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"

	DefaultOutputFormat = OutputFormatText
)

const (
	privateKeyFlag = "private-key"
	grpcURLFlag    = "kwil-provider"
	OutputFlag     = "output"

	privateKeyEnv = "private_key"
	grpcURLEnv    = "grpc_url"
	OutputKey     = "output"
)

// TODO: should i just use empty var??
var outputFormat = DefaultOutputFormat

func BindGlobalFlags(fs *pflag.FlagSet) {
	// Bind flags to environment variables
	fs.String(privateKeyFlag, "", "The private key of the wallet that will be used for signing")
	fs.String(grpcURLFlag, "", "The kwil provider endpoint")
	fs.Var(&outputFormat, OutputFlag, "the format for command output, either 'text' or 'json'")

	viper.BindPFlag(privateKeyEnv, fs.Lookup(privateKeyFlag))
	viper.BindPFlag(grpcURLEnv, fs.Lookup(grpcURLFlag))
	viper.BindPFlag(OutputKey, fs.Lookup(OutputFlag))
}

func GetOutputFormat() string {
	return viper.GetString(OutputKey)
}
