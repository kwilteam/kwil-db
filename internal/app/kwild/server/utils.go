package server

import (
	"context"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/pkg/engine"
	"github.com/kwilteam/kwil-db/pkg/extensions"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql"
	"github.com/kwilteam/kwil-db/pkg/sql/client"
)

var defaultFilePath string

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "/tmp"
	}

	defaultFilePath = fmt.Sprintf("%s/.kwil/sqlite/", dirname)
}

func init() {}

// connectExtensions connects to the provided extension urls.
func connectExtensions(ctx context.Context, urls []string) (map[string]*extensions.Extension, error) {
	exts := make(map[string]*extensions.Extension, len(urls))

	for _, url := range urls {
		ext := extensions.New(url)
		err := ext.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect extension '%s': %w", ext.Name(), err)
		}

		_, ok := exts[ext.Name()]
		if ok {
			return nil, fmt.Errorf("duplicate extension name: %s", ext.Name())
		}

		exts[ext.Name()] = ext
	}

	return exts, nil
}

func adaptExtensions(exts map[string]*extensions.Extension) map[string]engine.ExtensionInitializer {
	adapted := make(map[string]engine.ExtensionInitializer, len(exts))

	for name, ext := range exts {
		adapted[name] = extensionInitializeFunc(ext.CreateInstance)
	}

	return adapted
}

type extensionInitializeFunc func(ctx context.Context, metadata map[string]string) (*extensions.Instance, error)

func (e extensionInitializeFunc) CreateInstance(ctx context.Context, metadata map[string]string) (engine.ExtensionInstance, error) {
	return e(ctx, metadata)
}

type sqliteOpener struct {
	sqliteFilePath string
}

func newSqliteOpener(sqliteFilePath string) *sqliteOpener {
	if sqliteFilePath == "" {
		sqliteFilePath = defaultFilePath
	}

	return &sqliteOpener{
		sqliteFilePath: sqliteFilePath,
	}
}

func (s *sqliteOpener) Open(fileName string, logger log.Logger) (sql.Database, error) {
	return client.NewSqliteStore(fileName,
		client.WithLogger(logger),
		client.WithPath(s.sqliteFilePath),
	)
}
