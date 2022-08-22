package dba

import (
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/kwilteam/kwil-db/pkg/types/dba"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strings"
)

type Txn interface {
	Delete([]byte) error
	Get([]byte) ([]byte, error)
	Set([]byte, []byte) error
	Discard()
	Commit() error
}

type KVStore interface {
	Get([]byte) (string, error)
	Set([]byte, []byte) error
	Delete([]byte) error
	Exists([]byte) (bool, error)
	RunGC()
	DeleteByPrefix([]byte) error
	NewTransaction(bool) (Txn, error)
	GetAllByPrefix([]byte) ([][]byte, [][]byte, error)
	Close() error
}

type DBLoader struct {
	log    *zerolog.Logger
	Config *types.Config
	kv     KVStore
}

func New(conf *types.Config, kv KVStore) (*DBLoader, error) { // potentially returning an error in case we need more complex constructor later
	logger := log.With().Str("module", "dba").Int64("chainID", int64(conf.ClientChain.GetChainID())).Logger()

	go kv.RunGC()

	return &DBLoader{
		log:    &logger,
		Config: conf,
		kv:     kv,
	}, nil
}

// Make sure you pass this function a pointer to the DBConfig
func (d *DBLoader) LoadDatabase(dbConf dba.DatabaseConfig) error {
	d.log.Info().Msg("loading database")

	return nil
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
