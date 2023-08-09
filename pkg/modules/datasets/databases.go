package datasets

import "github.com/kwilteam/kwil-db/pkg/log"

type DatasetModule struct {
	engine       Engine
	accountStore AccountStore

	// feeMultiplier is the multiplier for the fee pricing
	// this is used to change / turn off transaction pricing
	feeMultiplier int64

	log log.Logger
}

func NewDatasetModule(engine Engine, accountStore AccountStore, opts ...DatabaseUseCaseOpt) *DatasetModule {
	d := &DatasetModule{
		engine:        engine,
		accountStore:  accountStore,
		feeMultiplier: 1,
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

type DatabaseUseCaseOpt func(*DatasetModule)

// WithFeeMultiplier sets the fee multiplier pricing
func WithFeeMultiplier(multiplier int64) DatabaseUseCaseOpt {
	return func(u *DatasetModule) {
		u.feeMultiplier = multiplier
	}
}

// WithLogger sets the logger
func WithLogger(logger log.Logger) DatabaseUseCaseOpt {
	return func(u *DatasetModule) {
		u.log = logger
	}
}
