package dba

import (
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/kwilteam/kwil-db/pkg/types/dba"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type KVStore interface {
	Get([]byte) ([]byte, error)
	Set([]byte, []byte) error
	Delete([]byte) error
	Exists([]byte) (bool, error)
	RunGC()
	DeleteByPrefix([]byte) error
	NewTransaction(bool) *badger.Txn
	GetAllByPrefix([]byte) ([][]byte, [][]byte, error)
	Close() error
}

type DBLoader struct {
	log    *zerolog.Logger
	Config *types.Config
	kv     KVStore
}

func NewLoader(conf *types.Config, kv KVStore) (*DBLoader, error) { // potentially returning an error in case we need more complex constructor later
	logger := log.With().Str("module", "dba").Int64("chainID", int64(conf.ClientChain.GetChainID())).Logger()

	go kv.RunGC()

	return &DBLoader{
		log:    &logger,
		Config: conf,
		kv:     kv,
	}, nil
}

func getDBPrefix(dbConf dba.DatabaseConfig) []byte {
	owner := dbConf.GetOwner()
	dbName := dbConf.GetName()

	sb := strings.Builder{}
	sb.WriteString(*owner)
	sb.WriteString("/")
	sb.WriteString(*dbName)

	return []byte(sb.String())
}
