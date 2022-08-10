package types

type Config struct {
	HTTP struct {
	}

	ClientChain ClientChain `json:"client_chain" mapstructure:"client_chain"`

	Log struct {
		Human bool `default:"false"`
		Debug bool `default:"false"`
	}
}
