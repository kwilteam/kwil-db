package types

type Config struct {
	HTTP struct {
	}

	ClientChain ClientChain `json:"client_chain" mapstructure:"client_chain"`
	Wallets     Wallets     `json:"wallets" mapstructure:"wallets"`
	Storage     Storage     `json:"storage" mapstructure:"storage"`
	Log         struct {
		Human bool `default:"false" json:"human" mapstructure:"human"`
		Debug bool `default:"false" mapstructure:"debug"`
	}
	Api  Api  `json:"api" mapstructure:"api"`
	Cost Cost `json:"cost" mapstructure:"cost"`
}

type Cost struct {
	Database DatabaseCosts `json:"database" mapstructure:"database"`
}

type Storage struct {
	Badger Badger `json:"badger" mapstructure:"badger"`
}

type Badger struct {
	Path string `json:"path" mapstructure:"path"`
}

type Api struct {
	Port        int `json:"port" mapstructure:"port"`
	TimeoutTime int `json:"timeout_time" mapstructure:"timeout_time"`
}

type DatabaseCosts struct {
	Create string `json:"create" mapstructure:"create"`
	Delete string `json:"delete" mapstructure:"delete"`
}
