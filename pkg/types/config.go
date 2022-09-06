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
	Api        Api             `json:"api" mapstructure:"api"`
	Cost       Cost            `json:"cost" mapstructure:"cost"`
	Auth       Auth            `json:"auth" mapstructure:"auth"`
	Friendlist []string        `json:"friends" mapstructure:"friends"`
	Friends    map[string]bool `json:"-" mapstructure:"-"`
	Peers      []string        `json:"peers" mapstructure:"peers"`
}

type Auth struct {
	ExpirationTime int `json:"token_expiration_time" mapstructure:"token_expiration_time"`
}

type Cost struct {
	Database DatabaseCosts `json:"database" mapstructure:"database"`
	Ddl      DDLCosts      `json:"ddl" mapstructure:"ddl"`
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

type DDLCosts struct {
	Table TableDDLCosts `json:"table" mapstructure:"table"`
	Role  RoleDDLCosts  `json:"role" mapstructure:"role"`
	Query QueryDDLCosts `json:"query" mapstructure:"query"`
}

type TableDDLCosts struct {
	Create string `json:"create" mapstructure:"create"`
	Delete string `json:"delete" mapstructure:"delete"`
	Modify string `json:"modify" mapstructure:"modify"`
}

type RoleDDLCosts struct {
	Create string `json:"create" mapstructure:"create"`
	Delete string `json:"delete" mapstructure:"delete"`
	Modify string `json:"modify" mapstructure:"modify"`
}

type QueryDDLCosts struct {
	Create string `json:"create" mapstructure:"create"`
	Delete string `json:"delete" mapstructure:"delete"`
}
