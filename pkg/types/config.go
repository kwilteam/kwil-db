package types

type Config struct {
	HTTP struct {
	}

	ClientChain ClientChain `json:"client_chain" mapstructure:"client_chain"`
	Wallets     Wallets     `json:"wallets" mapstructure:"wallets"`
	Storage     Storage     `json:"storage" mapstructure:"storage"`
	Log         struct {
		Human bool `default:"false"`
		Debug bool `default:"false" mapstructure:"debug"`
	}
}

type Storage struct {
	Badger Badger `json:"badger" mapstructure:"badger"`
}

type Badger struct {
	Path string `json:"path" mapstructure:"path"`
}
