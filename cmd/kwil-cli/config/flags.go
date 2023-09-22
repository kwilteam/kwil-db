package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cliCfg = DefaultKwilCliPersistedConfig()

const (
	defaultConfigDirName      = ".kwil_cli"
	defaultConfigFileName     = "config.json"
	AlternativeConfigHomePath = "/tmp"

	// NOTE: these flags below are also used as viper key names
	globalPrivateKeyFlag = "private-key"
	// globalProviderFlag historically there was a chain-provider flag,
	// we could/should change this flag to `provider`
	// also since the config file is using `grpc_url`, should change too
	// TODO: this is a breaking change
	globalProviderFlag = "kwil-provider"
	globalOutputFlag   = "output"
	globalTlsCertFlag  = "tls-cert-file"
	// NOTE: viper key name are used for viper related operations
	// here they are same `mapstructure` names defined in the config struct
	viperPrivateKeyName = "private_key"
	viperProviderName   = "grpc_url"
	viperTlsCertName    = "tls_cert_file"
	viperOutputName     = "output"
)

var defaultConfigFile string
var DefaultConfigDir string

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = AlternativeConfigHomePath
	}

	configPath := filepath.Join(dirname, defaultConfigDirName)
	DefaultConfigDir = configPath
	defaultConfigFile = filepath.Join(configPath, defaultConfigFileName)
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

var outputFormat = DefaultOutputFormat

func BindGlobalFlags(fs *pflag.FlagSet) {
	// Bind flags to environment variables
	fs.String(globalPrivateKeyFlag, cliCfg.PrivateKey, "The private key of the wallet that will be used for signing")
	fs.String(globalProviderFlag, cliCfg.GrpcURL, "The kwil provider endpoint")
	fs.String(globalTlsCertFlag, cliCfg.TLSCertFile, "The path to the TLS certificate, this is required if the kwil provider endpoint is using TLS")
	fs.Var(&outputFormat, globalOutputFlag, "the format for command output, either 'text' or 'json'")

	// Bind flags to viper, named by the flag name
	viper.BindPFlag(viperPrivateKeyName, fs.Lookup(globalPrivateKeyFlag))
	viper.BindPFlag(viperProviderName, fs.Lookup(globalProviderFlag))
	viper.BindPFlag(viperTlsCertName, fs.Lookup(globalTlsCertFlag))
	viper.BindPFlag(viperOutputName, fs.Lookup(globalOutputFlag))
}

func GetOutputFormat() string {
	return outputFormat.String()
}
