package processing

import (
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type EventProcessor struct {
	log       zerolog.Logger
	eventChan chan map[string]interface{}
}

func New(conf *types.Config, ech chan map[string]interface{}) *EventProcessor {
	logger := log.With().Str("module", "processing").Int64("chainID", int64(conf.ClientChain.GetChainID())).Logger()
	return &EventProcessor{
		log:       logger,
		eventChan: ech,
	}
}
